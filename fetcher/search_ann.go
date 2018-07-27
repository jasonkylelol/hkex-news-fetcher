/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements. See the NOTICE file
distributed with this work for additional information
regarding copyright ownership. The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License. You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied. See the License for the
specific language governing permissions and limitations
under the License.
*/

/*
author: github.com/jasonkylelol
date: 2018-07-20
*/

package fetcher

import (
	"crypto/sha1"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func SearchSpecificAnn(stock string, start_timestamp, end_timestamp int64, infos *[]AnnInfo) error {
	content, err := urlQuery()
	if err != nil {
		return fmt.Errorf("urlQuery failed for [%s]", err)
	}
	formNode, err := formNodeParse(content)
	if err != nil {
		return fmt.Errorf("formNodeParse failed for [%s]\n", err)
	}
	vals, err := viewStateParse(formNode)
	if err != nil {
		return fmt.Errorf("viewStateParse failed for [%s]\n", err)
	}
	err = formRequest(vals, stock)
	if err != nil {
		return fmt.Errorf("formRequest failed for [%s]\n", err)
	}
	return nil
}

func urlQuery() (string, error) {
	queryUrl := "http://www.hkexnews.hk/listedco/listconews/advancedsearch/search_active_main_c.aspx"
	resp, err := http.Get(queryUrl)
	if err != nil {
		fmt.Printf("http url[%s] failed for err[%s]\n", queryUrl, err)
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf("http url[%s] with status code[%d]\n", queryUrl, resp.StatusCode)
		return "", fmt.Errorf("http resp status code[%d]\n", resp.StatusCode)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("response read failed for [%s]\n", err)
		return "", err
	}
	return string(content), nil
}

func formNodeParse(content string) (*goquery.Selection, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		fmt.Printf("goquery init doc failed for [%s]\n", err)
		return nil, err
	}
	formNode := doc.Find("form")

	attrName, _ := formNode.Attr("name")
	attrMethod, _ := formNode.Attr("method")
	attrAction, _ := formNode.Attr("action")
	attrId, _ := formNode.Attr("id")
	fmt.Printf("formNode attr: name[%s] method[%s] action[%s] id[%s]\n",
		attrName, attrMethod, attrAction, attrId)

	// formHtml, _ := formNode.Html()
	// fmt.Printf("formNode html\n%s\n", formHtml)

	return formNode, nil
}

func viewStateParse(node *goquery.Selection) (map[string]string, error) {
	// formHtml, _ := node.Html()
	// fmt.Printf("HTML is \n%s\n", formHtml)

	vals := map[string]string{}

	inputNodes := node.ChildrenFiltered("input")
	inputNodes.Each(func(i int, s *goquery.Selection) {
		attrName, _ := s.Attr("name")
		// attrId, _ := s.Attr("id")
		attrValue, _ := s.Attr("value")
		// fmt.Printf("inputNode attr: name[%s] id[%s]\n", attrName, attrId)
		vals[attrName] = attrValue
	})
	return vals, nil
}

func formRequest(vals map[string]string, stock string) error {
	viewState := vals["__VIEWSTATE"]
	viewStateGen := vals["__VIEWSTATEGENERATOR"]
	txtToday := vals["ctl00$txt_today"]
	hfStatus := vals["ctl00$hfStatus"]

	queryUrl := "http://www.hkexnews.hk/listedco/listconews/advancedsearch/search_active_main_c.aspx"

	var fromDate, toDate string
	var fromDay, fromMon, fromYear int
	var toDay, toMon, toYear int
	_, err := fmt.Sscanf(fromDate, "%d-%d-%d", &fromYear, &fromMon, &fromDay)
	if err != nil {
		fmt.Printf("Sscanf FromDate[%s] failed for [%s]\n", fromDate, err)
		return err
	}
	_, err = fmt.Sscanf(toDate, "%d-%d-%d", &toYear, &toMon, &toDay)
	if err != nil {
		fmt.Printf("Sscanf ToDate[%s] failed for [%s]\n", toDate, err)
		return err
	}

	form := url.Values{}
	form.Set("__VIEWSTATE", viewState)
	form.Set("__VIEWSTATEGENERATOR", viewStateGen)
	form.Set("__VIEWSTATEENCRYPTED", "")
	form.Set("ctl00$txt_today", txtToday)
	form.Set("ctl00$hfStatus", hfStatus)
	form.Set("ctl00$hfAlert", "")
	form.Set("ctl00$txt_stock_code", stock)
	form.Set("ctl00$txt_stock_name", "")
	form.Set("ctl00$rdo_SelectDocType", "rbAll")
	form.Set("ctl00$sel_tier_1", "-2")
	form.Set("ctl00$sel_DocTypePrior2006", "-1")
	form.Set("ctl00$sel_tier_2_group", "-2")
	form.Set("ctl00$sel_tier_2", "-2")
	form.Set("ctl00$ddlTierTwo", "59,1,7")
	form.Set("ctl00$ddlTierTwoGroup", "26,5")
	form.Set("ctl00$txtKeyWord", "")
	form.Set("ctl00$rdo_SelectDateOfRelease", "rbManualRange")
	form.Set("ctl00$sel_DateOfReleaseFrom_d", fmt.Sprintf("%02d", fromDay))
	form.Set("ctl00$sel_DateOfReleaseFrom_m", fmt.Sprintf("%02d", fromMon))
	form.Set("ctl00$sel_DateOfReleaseFrom_y", fmt.Sprintf("%04d", fromYear))
	form.Set("ctl00$sel_DateOfReleaseTo_d", fmt.Sprintf("%02d", toDay))
	form.Set("ctl00$sel_DateOfReleaseTo_m", fmt.Sprintf("%02d", toMon))
	form.Set("ctl00$sel_DateOfReleaseTo_y", fmt.Sprintf("%04d", toYear))
	form.Set("ctl00$sel_defaultDateRange", "SevenDays")
	form.Set("ctl00$rdo_SelectSortBy", "rbDateTime")

	// fmt.Printf("post form values\n%+v\n", form)

	resp, err := http.PostForm(queryUrl, form)
	if err != nil {
		fmt.Printf("http form post failed for [%s]\n", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf("http post with status code[%d]\n", resp.StatusCode)
		return fmt.Errorf("http resp status code[%d]\n", resp.StatusCode)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("response read failed for [%s]\n", err)
		return err
	}
	fmt.Printf("request for stock[%s] resp len[%d]\n", stock, len(string(content)))
	err = searchResultParse(string(content))
	if err != nil {
		fmt.Printf("searchResultParse failed for [%s]\n", err)
	}
	return nil
}

func searchResultParse(content string) error {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		fmt.Printf("goquery init doc failed for [%s]\n", err)
		return err
	}

	formNode := doc.Find("form")

	vals, err := viewStateParse(formNode)
	if err != nil {
		fmt.Printf("viewStateParse failed for [%s]\n", err)
	}

	attrName, _ := formNode.Attr("name")
	attrMethod, _ := formNode.Attr("method")
	attrAction, _ := formNode.Attr("action")
	attrId, _ := formNode.Attr("id")
	fmt.Printf("formNode attr: name[%s] method[%s] action[%s] id[%s]\n",
		attrName, attrMethod, attrAction, attrId)
	// fmt.Printf("view state [%+v]\n", vals)
	// formHtml, _ := formNode.Html()
	// fmt.Printf("formNode html \n%s\n", formHtml)

	trNodes := formNode.ChildrenFiltered("table").Eq(2).ChildrenFiltered("tbody").ChildrenFiltered("tr")
	navigateNode := trNodes.Eq(2)
	annNodes := trNodes.Eq(1).Find("tbody")
	/*
		navigateNodeHtml, _ := navigateNode.Html()
		annNodesHtml, _ := annNodes.Html()
		fmt.Printf("navigateNodeHtml is \n%s\nannNodesHtml is \n%s\n", navigateNodeHtml, annNodesHtml)
	*/
	overlap_timestamp, err := annResultParse(annNodes)
	if err != nil {
		fmt.Printf("annResultParse failed for [%s]\n", err)
		return err
	}
	var fromDate string
	var year, mon, day int
	// 2018-07-04
	_, err = fmt.Sscanf(fromDate, "%d-%d-%d", &year, &mon, &day)
	if err != nil {
		fmt.Printf("parse FromDate[%d] failed for [%s]\n", fromDate, err)
		return err
	}
	from_timestamp := time.Date(year, time.Month(mon), day, 0, 0, 0, 0, time.UTC).Unix()
	if overlap_timestamp < from_timestamp {
		fmt.Printf("range stoped at [%s]\n", time.Unix(overlap_timestamp, 0).String())
		return nil
	}

	navigateNodeHtml, _ := navigateNode.Html()
	fmt.Printf("\nnavigateNodeHtml is:\n%s\n", navigateNodeHtml)
	navigateNode.Find("td").Each(func(i int, s *goquery.Selection) {
		if s.AttrOr("align", "") != "right" {
			return
		}
		time.Sleep(4 * time.Second)
		s.Find("input").Each(func(i int, ss *goquery.Selection) {
			if strings.Contains(ss.AttrOr("name", ""), "Next") {
				err := nextFormRequest(vals)
				if err != nil {
					fmt.Printf("nextFormRequest failed for [%s]\n")
					return
				}
			}
		})
	})
	return nil
}

func nextFormRequest(vals map[string]string) error {
	viewState := vals["__VIEWSTATE"]
	viewStateGen := vals["__VIEWSTATEGENERATOR"]
	// txtToday := vals["ctl00$txt_today"]
	// hfStatus := vals["ctl00$hfStatus"]

	stock := "00763"
	queryUrl := "http://www.hkexnews.hk/listedco/listconews/advancedsearch/search_active_main_c.aspx"

	form := url.Values{}
	form.Set("__VIEWSTATE", viewState)
	form.Set("__VIEWSTATEGENERATOR", viewStateGen)
	form.Set("__VIEWSTATEENCRYPTED", "")
	form.Set("ctl00$btnNext2.x", "22")
	form.Set("ctl00$btnNext2.y", "8")

	resp, err := http.PostForm(queryUrl, form)
	if err != nil {
		fmt.Printf("http form post failed for [%s]\n", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		fmt.Printf("http post with status code[%d]\n", resp.StatusCode)
		return fmt.Errorf("http resp status code[%d]\n", resp.StatusCode)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("response read failed for [%s]\n", err)
		return err
	}
	fmt.Printf("request for stock[%s] resp len[%d]\n", stock, len(string(content)))
	err = searchResultParse(string(content))
	if err != nil {
		fmt.Printf("searchResultParse failed for [%s]\n", err)
	}
	return nil
}

func annResultParse(node *goquery.Selection) (int64, error) {
	// nodeHtml, _ := node.Html()
	// fmt.Printf("nodeHtml is \n%s\n", nodeHtml)
	overlap_timestamp := time.Now().Unix()
	node.ChildrenFiltered("tr").Each(func(i int, s *goquery.Selection) {
		_, exist := s.Attr("bgcolor")
		if !exist {
			return
		}
		tdNodes := s.ChildrenFiltered("td")
		tdNodes.Eq(0).Find("span").Find("br").AfterHtml("-")
		var year, mon, day, hour, min int
		// 21/05/2018-16:31
		_, err := fmt.Sscanf(tdNodes.Eq(0).Find("span").Text(),
			"%2d/%2d/%4d-%2d:%2d", &day, &mon, &year, &hour, &min)
		if err != nil {
			fmt.Printf("sscanf str[%s] failed for error[%s]\n",
				tdNodes.Eq(0).Find("span").Text(), err)
			return
		}
		timestamp := time.Date(year, time.Month(mon), day, hour, min, 0, 0, time.UTC).Unix()
		ann_time := tdNodes.Eq(0).Find("span").Text()
		ann_stock := tdNodes.Eq(1).Find("span").Text() + ".HK"
		ann_name := tdNodes.Eq(2).Find("span").Text()
		ann_title := tdNodes.Eq(3).Find("span").Eq(0).Text()
		ann_title_ext := tdNodes.Eq(3).Find("a").Text()
		ann_link := "http://www.hkexnews.hk" + tdNodes.Eq(3).Find("a").AttrOr("href", "nil")
		guid := fmt.Sprintf("%x", sha1.Sum([]byte(ann_link)))
		fmt.Printf("ann_time[%s] ann_stock[%s] ann_name[%s] ann_title[%s] ann_title_ext[%s] ann_link[%s] timestamp[%d] guid[%s]\n",
			ann_time, ann_stock, ann_name, ann_title, ann_title_ext, ann_link, timestamp, guid)
		overlap_timestamp = timestamp

		ann_info := AnnInfo{}
		// TODO(fill AnnouncementInfo)
		_ = ann_info
	})
	return overlap_timestamp, nil
}
