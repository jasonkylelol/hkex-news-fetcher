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
	"strings"
	"time"
)

func httpRequest() (string, error) {
	/*
		req, err := http.NewRequest("GET", "http://www.hkexnews.hk", nil)
		if err != nil {
			glog.Errorf("http new request catch an error[%s]", err)
			return
		}
		// req.Header.Add("Referer", `http://www.hkexnews.hk/listedco/listconews/mainindex/SEHK_LISTEDCO_DATETIME_TODAY_C.HTM`)
		req.Header.Add("User-Agent", `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.181 Safari/537.36`)
		req.Header.Add("Accept", `text/html`)
		req.Header.Add("Accept-Language", `zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7`)
		// req.Header.Add("Host", `www.hkexnews.hk`)
		// req.Header.Add("Connection", `keep-alive`)
		var resp *http.Response
		client := &http.Client{
			Transport:     nil,
			CheckRedirect: nil,
			Jar:           nil,
			Timeout:       5 * time.Second,
		}
		resp, err = client.Do(req)
	*/
	url := "http://www.hkexnews.hk/listedco/listconews/mainindex/SEHK_LISTEDCO_DATETIME_TODAY_C.HTM"
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("http request[%s] catch an error[%s]", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("http request[%s] with status code[%d]", url, resp.StatusCode)
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("http request[%s] read body catch an error[%s]", url, err)
	}
	return string(content), nil
}

func FetchLatestAnn(infos *[]AnnInfo) error {
	if infos == nil {
		return fmt.Errorf("invalid nil parameter")
	}
	content, err := httpRequest()
	if err != nil {
		return fmt.Errorf("fetch latest failed for http request[%s]", err)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return fmt.Errorf("goquery init doc failed for [%s]", err)
	}
	doc.Find(".row0,.row1").Each(func(i int, s *goquery.Selection) {
		s.Find("td").Eq(0).Find("br").AfterHtml("-")
		time_str := s.Find("td").Eq(0).Text()
		var year, mon, day, hour, min int
		// format: 21/05/2018-16:31
		_, err = fmt.Sscanf(time_str, "%2d/%2d/%4d-%2d:%2d", &day, &mon, &year, &hour, &min)
		if err != nil {
			fmt.Printf("sscanf time failed for [%s]", err)
			return
		}
		timestamp := time.Date(year, time.Month(mon), day, hour, min, 0, 0, time.UTC).Unix()
		stock_id := s.Find("td").Eq(1).Text()
		stock_name := s.Find("td").Eq(2).Find("nobr").Text()
		var ann_link string
		href, exist := s.Find("a.news").Attr("href")
		if exist {
			ann_link = "http://www.hkexnews.hk" + href
		}
		ann_link_title := s.Find("a.news").Text()
		ann_text := s.Find("td").Eq(3).Find("div").Text()
		ann_link_title = strings.Replace(ann_link_title, "\n", "", -1)
		ann_link_title = strings.Replace(ann_link_title, " ", "", -1)
		ann_text = strings.Replace(ann_text, "\n", "", -1)
		ann_text = strings.Replace(ann_text, " ", "", -1)
		guid := fmt.Sprintf("%x", sha1.Sum([]byte(ann_link)))
		info := AnnInfo{
			Guid:      guid,
			FileTitle: ann_link_title,
			FileLink:  ann_link,
			StockID:   stock_id,
			StockName: stock_name,
			TimeStr:   time_str,
			Timestamp: timestamp,
			Content:   ann_text,
		}
		*infos = append(*infos, info)
	})
	return nil
}
