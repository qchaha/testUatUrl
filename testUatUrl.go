package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func requestWithCookie(api string, method string, data io.Reader) string {
	req, err := http.NewRequest(method, api, data)

	if err != nil {
		fmt.Println("create new request error...")
		fmt.Println(err)
		return "create new request error..."
	}
	// uuid must be configurated, other optional
	req.AddCookie(&http.Cookie{Name: "XSRF-TOKEN", Value: "eyJpdiI6IkdTelR0SmxtMmZ5OWFyTzBOcUM2cEE9PSIsInZhbHVlIjoiRGdMMXNlTVB4eXVhQ3BTc0cyaTJuQ2sxZnQyWndQRUU2VEppZ2VYbzJrVlZCcXRmdVBoODVpWTBndnUzamp4bnROTEtFUTVTV2EzbVdRVVB4U2Q4SXc9PSIsIm1hYyI6IjQzYTJiYzI4OTQ3OGQ5MGQwNTM5MzA4Njc1OTMwM2UyZmM4NDc1MzBkNjU2ZDk3MzMyNzk2MGU2ZGI2NDVlNjYifQ%3D%3D"})
	req.AddCookie(&http.Cookie{Name: "laravel_session", Value: "eyJpdiI6IlcxRnRla0dWKzc1b3BZWnZxS3BcL3F3PT0iLCJ2YWx1ZSI6InBNUDRQaE4yb2JuZlpLdzBiRXVwSm1xdkhlSHRNbnI3aEFtVmlDTlZYY1NEQWhka0RUMWp6U1pHS2dHVHdsM1JEeUtldWlBRTZBa2dFVElwRFh0dTh3PT0iLCJtYWMiOiJkMTRiYTU0ZDYyYmQ3YWU4ZmI4NTk5NDYzMzU1ZjI1ZDM3ODFjZjZiNGJmM2EyZjQzYjNkYmM4YzJjMzA1YjUzIn0%3D"})
	req.AddCookie(&http.Cookie{Name: "region", Value: "CN"})
	req.AddCookie(&http.Cookie{Name: "f", Value: "banner_fordeal_yantong_1207%7C%7C2018-12-13"})
	req.AddCookie(&http.Cookie{Name: "lan", Value: "zh"})
	req.AddCookie(&http.Cookie{Name: "timezone", Value: "%2B8"})
	req.AddCookie(&http.Cookie{Name: "_ga", Value: "GA1.2.1293514611.1544335822"})
	req.AddCookie(&http.Cookie{Name: "_gid", Value: "GA1.2.836380822.1544689259"})
	req.AddCookie(&http.Cookie{Name: "build", Value: "h5"})
	req.AddCookie(&http.Cookie{Name: "cur", Value: "CNY"})
	req.AddCookie(&http.Cookie{Name: "uuid", Value: "7142387b7e764bd6802be4aedc4e1dba"})
	req.AddCookie(&http.Cookie{Name: "version", Value: "h5"})
	req.AddCookie(&http.Cookie{Name: "__cfduid", Value: "d7f4261384761bbaeae5ef5aaca78c2431544267489"})

	if method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	client := http.Client{
		Timeout: time.Second * 4}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Println("request error")
		fmt.Println(err)
		return "request error..."
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		//fmt.Println(api + "\nresp.StatusCode: " + strconv.Itoa(resp.StatusCode))
		return api + "\nerror...\nresp.StatusCode: " + strconv.Itoa(resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)

	return string(body)
}

func main() {
	rHttpGet, _ := regexp.Compile("^http")
	rHttpPost, _ := regexp.Compile("\\[\\[\\[.+?\\]\\]\\]")
	rPostData, _ := regexp.Compile("\\(.+?\\)")
	rData, _ := regexp.Compile("\"(.+?)\",\"(.+?)\"")
	rReturnCode, _ := regexp.Compile("\"code\":\"?\\d+\"?")
	rFindCid, _ := regexp.Compile("\"cid\":(\\d+)")
	replace := strings.NewReplacer("[", "", "]", "")
	var responseBody, returnCode, cid string
	data := url.Values{}

	// post request
	file, err := os.Open("./postUrls.txt")

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if rHttpPost.FindAllString(scanner.Text(), -1) != nil {
			// [[[http://uat.duolainc.com/api3/order/create]]][[[("address_id","1648729")("coupon_type","coupon")("id","39287170")]]]
			api := replace.Replace(rHttpPost.FindAllString(scanner.Text(), -1)[0])
			// first[[[*]]]
			if strings.Index(api, "/api2/cart/del/") != -1 {
				api = api + cid
			}
			reqMsg := fmt.Sprintf("POST %s", api)
			fmt.Println(reqMsg)
			// second[[[]]]
			if len(rHttpPost.FindAllString(scanner.Text(), -1)) > 1 {
				parameters := replace.Replace(rHttpPost.FindAllString(scanner.Text(), -1)[1])
				// ("address_id","1648729")("coupon_type","coupon")("id","39287170")
				for i := 0; i < len(rPostData.FindAllString(parameters, -1)); i++ {
					// "address_id","1648729"
					dataGroup := rData.FindAllStringSubmatch(rPostData.FindAllString(parameters, -1)[i], -1)
					for _, d := range dataGroup {
						//fmt.Println(d[1] + "  " + d[2])
						data.Add(d[1], d[2])
					}
				}
			}
			responseBody = requestWithCookie(api, "POST", strings.NewReader(data.Encode()))
			// filter error request
			if strings.Index(responseBody, "error") != -1 {
				fmt.Println(responseBody)
				os.Exit(4)
			}
			returnCode = rReturnCode.FindString(responseBody)
			// save cid which will be used in cart deleting
			if strings.Index(api, "/cart/add") != -1 {
				match := rFindCid.FindAllStringSubmatch(responseBody, -1)
				for _, tcid := range match {
					cid = tcid[1]
				}
			}
			fmt.Println(returnCode + "\n")
			// filter logical error
			if strings.Index(returnCode, "1001") == -1 {
				fmt.Println("code is not equal 1001...")
				os.Exit(5)
			}
		} else {
			fmt.Println(scanner.Text())
		}
	}
	file.Close()
	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	// get request
	file, err = os.Open("./getUrls.txt")
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
	}
	defer file.Close()
	scanner = bufio.NewScanner(file)
	for scanner.Scan() {
		if rHttpGet.MatchString(scanner.Text()) == true {
			regMsg := fmt.Sprintf("GET %s", scanner.Text())
			fmt.Println(regMsg)
			responseBody = requestWithCookie(scanner.Text(), "GET", nil)
			returnCode = rReturnCode.FindString(responseBody)
			fmt.Println(returnCode + "\n")
			// filter error request
			if strings.Index(responseBody, "error") != -1 {
				fmt.Println(responseBody)
				os.Exit(4)
			}
		} else {
			fmt.Println(scanner.Text())
		}
	}
}
