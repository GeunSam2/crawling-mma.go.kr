import requests
from bs4 import BeautifulSoup as bsp
import time
import pandas as pd

URL = "http://work.mma.go.kr/caisBYIS/search/byjjecgeomsaek.do"

s1 = requests.Session()
max_page = 2 #특정 키워드로 검색했을때, 출력되는 게시판 페이지 리스트에서 크롤링할 페이지수를 지정
# 한 페이지당 10개 업체가 등록되어 있음

page_list = []
print ("업체리스트를 조회 중입니다. 잠시만 기다려주세요...")
for page_num in range(1,max_page):
    data = {
        "al_eopjong_gbcd":"11111,11112",
        "bjinwonym":"",
        "chaeyongym":"",
        "eopche_nm":"",
        "eopjong_gbcd":"1",
        "eopjong_gbcd_list":"11111,11112",
        "gegyumo_cd":"",
        "juso":"",
        "menu_id":"",
        "pageIndex":page_num,
        "pageUnit":"10",
        "searchCondition":"",
        "searchKeyword":"",
        "sido_addr":"서울특별시",
        "sigungu_addr":""
        }
    time.sleep(0.5)
    req1 = s1.post(URL,data=data)
    req1.encoding = None
    time.sleep(0.5)
    soup = bsp(req1.text,'html5lib')

    for i in soup.find_all("th","title t-alignLt pl20px") :
        comp = str(i).split("byjjeopche_cd=")[1].split("&")[0]
        page_list.append(comp)
print ()
print ("총 업체수 : "+str(len(page_list)))




dic = {}
print ()
print ("업체 세부 정보 수집을 시작합니다. 잠시만 기다려 주십시오...")
for i in page_list :
    URL2 = "https://work.mma.go.kr/caisBYIS/search/byjjecgeomsaekView.do?menu_id=m_m6&pageIndex=1&byjjeopche_cd="+str(i)+"&eopjong_gbcd=1&gegyumo_cd=&eopche_nm=&sido_addr=%EC%84%9C%EC%9A%B8%ED%8A%B9%EB%B3%84%EC%8B%9C&sigungu_addr=&chaeyongym=&bjinwonym=&eopjong_gbcd_list=11111,11112"

    req2 = s1.get(URL2)
    req2.encoding = None
    soup2 = bsp(req2.text,"html5lib")
    comp2 = []
    for j in soup2.find_all("td")[:4] :
        comp2.append(str(j).split(">")[1].split("</")[0])
    dic.update({comp2[0]:[]})
    for k in (comp2[1:]) :
         dic[comp2[0]].append(k)


# 수집한 데이터를 데이터 프레임 형태로 제작하기 위해 재정렬
c_name=[]
c_addres=[]
c_call=[]
c_fax=[]
for i in dic :
    c_name.append(i)
    c_addres.append(dic[i][0])
    c_call.append(dic[i][1])
    c_fax.append(dic[i][2])

#DataFrame 제작
data_set={'회사명':c_name,
          '회사 주소':c_addres,
          '회사 연락처':c_call,
          '회사 팩스주소':c_fax
          }
df1=pd.DataFrame(data=data_set)

#저장할 파일 이름 입력
print ("업체정보 수집이 완료되었습니다.")
filename=input("저장할 엑셀 파일 이름을 입력해주세요(확장자 없이) : ")
filename=filename+".xlsx"

#엑셀 저장 엔진을 생성
writer=pd.ExcelWriter(filename, engine='xlsxwriter')

df1.to_excel(writer, sheet_name="병역업체정보")
writer.save()

print ("저장이 완료되었습니다.")
