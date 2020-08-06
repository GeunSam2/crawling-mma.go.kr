package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/tealeg/xlsx"
)

const url string = "http://work.mma.go.kr/caisBYIS/search/byjjecgeomsaek.do"

type extractedJob struct {
	name string
	addr string
	call string
	fax  string
}

var params = map[string]string{
	"al_eopjong_gbcd":   "11111,11112",
	"bjinwonym":         "",
	"chaeyongym":        "",
	"eopche_nm":         "",
	"eopjong_gbcd":      "1",
	"eopjong_gbcd_list": "11111,11112",
	"gegyumo_cd":        "",
	"juso":              "",
	"menu_id":           "",
	"pageIndex":         "0",
	"pageUnit":          "10",
	"searchCondition":   "",
	"searchKeyword":     "",
	"sido_addr":         "서울특별시",
	"sigungu_addr":      "",
}

func main() {
	maxPage := getPages()
	fmt.Println("TotalPage : ", maxPage)
	fmt.Println("업체리스트를 조회 중입니다. 잠시만 기다려주세요...")
	var ids []string
	var datas []extractedJob
	c := make(chan []string)
	for i := 0; i < maxPage; i++ {
		go getPage(i, c)
	}
	for i := 0; i < maxPage; i++ {
		ids = append(ids, (<-c)...)
	}
	fmt.Println("총 업체수 : ", len(ids))
	print("업체 세부 정보 수집을 시작합니다. 잠시만 기다려 주십시오...")
	cData := make(chan extractedJob, 1)
	for _, id := range ids {
		go getCompanyInfo(id, cData)
	}
	for i := 0; i < len(ids); i++ {
		datas = append(datas, <-cData)
	}
	fmt.Println("업체정보 수집이 완료되었습니다.")
	writeExcel(datas)

}

func writeExcel(datas []extractedJob) {
	var filename string
	fmt.Print("저장할 엑셀 파일 이름을 입력해주세요(확장자 없이) : ")
	fmt.Scanln(&filename)
	filename += ".xlsx"
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	checkErr(err)
	row := sheet.AddRow()
	for i := 0; i < 4; i++ {
		cell := row.AddCell()
		switch i {
		case 0:
			cell.Value = "회사명"
		case 1:
			cell.Value = "회사 주소"
		case 2:
			cell.Value = "회사 연락처"
		case 3:
			cell.Value = "회사 팩스주소"
		}
	}
	for _, data := range datas {
		appendExtractedJob(data, sheet)
	}
	checkErr(file.Save(filename))
	fmt.Println("저장이 완료되었습니다.")
}

func appendExtractedJob(data extractedJob, sheet *xlsx.Sheet) {
	row := sheet.AddRow()
	for i := 0; i < 4; i++ {
		cell := row.AddCell()
		switch i {
		case 0:
			cell.Value = data.name
		case 1:
			cell.Value = data.addr
		case 2:
			cell.Value = data.call
		case 3:
			cell.Value = data.fax
		}
	}
}

func makeURLParams(url string, params map[string]string) string {
	url += "?"
	for key, val := range params {
		url += key + "=" + val + "&"
	}
	return url
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkStatus(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status:", res.StatusCode)
	}
}

func getCompanyInfo(id string, c chan<- extractedJob) {
	URL := "https://work.mma.go.kr/caisBYIS/search/byjjecgeomsaekView.do?byjjeopche_cd=" + id
	var data extractedJob
	res, err := http.Get(URL)
	checkErr(err)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
	doc.Find("td").Each(func(i int, s *goquery.Selection) {
		switch i {
		case 0:
			data.name = s.Text()
		case 1:
			data.addr = s.Text()
		case 2:
			data.call = s.Text()
		case 3:
			data.fax = s.Text()
		default:
			return
		}
	})
	c <- data
}

func getPage(i int, c chan<- []string) {
	var ids []string
	params["pageIndex"] = strconv.Itoa(i)
	res, err := http.Get(makeURLParams(url, params))
	checkErr(err)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
	doc.Find(".brd_list_n>tbody").Find("th").Each(func(i int, s *goquery.Selection) {
		id, exits := s.Find(".title.t-alignLt.pl20px>a").Attr("href")
		if exits {
			ids = append(ids, strings.Split(strings.Split(id, "byjjeopche_cd=")[1], "&")[0])
		}
	})
	c <- ids
}

func getPages() int {
	res, err := http.Get(makeURLParams(url, params))
	checkErr(err)
	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)
	return (doc.Find(".page_move_n>a").Length())
}
