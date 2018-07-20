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

package main

import (
	"fmt"
	"github.com/jasonkylelol/hkex_news_fetcher"
)

func main() {
	// 1. test LatestAnn
	infos, err := hkex_news_fetcher.LatestAnn()
	if err != nil {
		fmt.Printf("LatestAnn failed for [%s]\n", err)
		return
	}
	for _, v := range infos {
		fmt.Printf("%+v\n", v)
	}

	// 2. test SearchAnn
}
