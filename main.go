package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	g "github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/PonyWilliam/go-arcsoft/RfidUtils"
	. "github.com/windosx/face-engine/v4"
	"github.com/windosx/face-engine/v4/util"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)
type Obj struct{
	name string
	nums string
	id int64
	score int64
	Image util.ImageInfo
	FaceInfos MultiFaceInfo
	FaceInfo SingleFaceInfo
	result FaceFeature
}
type res struct{
	Code int `json:"code"`
	Msg string `json:"msg"`
	Token string `json:"token"`
}
type res2 struct{
	Code int `json:"code"`
	Data struct{Workers []Msg `json:"workers"`} `json:"data"`
}
type res3 struct{
	ID int64 `json:"id"`
	Name string `json:"product_name"`
}
type allRes struct{
	Code int `json:"code"`
	Msg string `json:"msg"`
}
type Devices struct{
	device []res3
}
type Msg struct {
	ID int `json:"ID"`
	Name string `json:"Name"`
	Nums string `json:"Nums"`
	Score int `json:"Score"`
	Telephone string `json:"Telephone"`
}
var (
	engine *FaceEngine
	window *gocv.Window
	media  *gocv.VideoCapture
	maxindex int
	preResult bool
	baseurl string
	dataurl string
	localPath string
	token string
	title imgui.Font
	objs []Obj
	Rfid [][]byte
	count int
	str []interface{}
	devices Devices
	ok bool
	now string
	AllRes allRes
	wnd *g.MasterWindow
	showWindow bool
)
func DownloadImage(nums string)error{
	//根据nums下载图片接口
	DownLoadUrl := baseurl + nums + ".png?time=" + strconv.FormatInt(time.Now().Unix(),10)
	fmt.Println(DownLoadUrl)
	resp,err := http.Get(DownLoadUrl)
	if err!= nil {
		return err
	}
	body,err := ioutil.ReadAll(resp.Body)
	if err!= nil {
		return err
	}
	out,err := os.Create(localPath + nums + ".png")
	if err!= nil {
		return err
	}
	_, err = io.Copy(out, bytes.NewBuffer(body))
	if err!= nil {
		return err
	}
	return nil
}
func getToken(){
	fmt.Println(123)
	val := url.Values{}
	val.Set("username","admin")
	val.Set("password","admin")
	resp,err := http.PostForm(fmt.Sprintf("%swork/login", dataurl),val)
	if err != nil{
		log.Fatal("error in login")
	}
	defer resp.Body.Close()
	bs,err := ioutil.ReadAll(resp.Body)
	if err != nil{
		log.Fatal(err)
	}
	data := &res{}
	_ = json.Unmarshal(bs, &data)
	fmt.Println(data)
	token = data.Token
}
func postBorrow(pid int64)(string,error){
	new_pid := strconv.FormatInt(pid,10)
	val := url.Values{}
	val.Set("wid","admin")
	val.Set("password","admin")
	resp,err := http.PostForm(fmt.Sprintf("%swork/borrow/%s",dataurl, new_pid),val)
	if err != nil{
		return "", err
	}
	defer resp.Body.Close()
	bs,err := ioutil.ReadAll(resp.Body)
	if err != nil{
		return "", err
	}
	data := &allRes{}
	_ = json.Unmarshal(bs, &data)
	if data == nil || data.Code!=200 {
		return data.Msg,errors.New("返回出错")
	}
	return data.Msg,nil
}
func init() {
	//1. 激活虹软
	err := OnlineActivation("8tM7EeBHZhL1De6wgRs8nJEJkoxy96VSKAMypTSeY7By", "F4V8HBCEYwsm4EU3XifvWU6VbGRhDbmkSAuibdmqTSUv", "8691-116F-H133-TE67")
	fmt.Println(123)
	if err != nil {
		panic(err)
	}
	fmt.Println(123)
	//2. 设置变量以及登录后台获取数据库信息
	baseurl = "http://arcsoft.dadiqq.cn/face/" //初始化获取图片的地址
	dataurl = "http://192.168.97.209:8080/"     //初始化数据接口
	localPath = "C:\\faces\\"
	getToken()
}
func GetFiles(){
	//1.数据库拉取员工
	var err error
	client := &http.Client{}
	request,err := http.NewRequest("GET",fmt.Sprintf("%swork/workers", dataurl),nil)
	if err != nil{
		log.Fatal(err)
	}
	request.Header.Add("Authorization", token) //携带token访问
	temp,_ := client.Do(request)
	response,err := ioutil.ReadAll(temp.Body)
	if err != nil{
		log.Fatal(err)
	}
	fmt.Println(string(response))
	res2 := &res2{}
	err = json.Unmarshal(response, &res2)

	if err != nil{
		log.Fatal(err)
	}
	defer temp.Body.Close()
	//2. 从阿里云oss拉取图片并保存到本地(release环境下理论说只需要启动一次，所以不考虑下载问题)
	if res2.Data.Workers == nil {
		panic("没有需要对比的员工")
	}
	for _,temp2 := range res2.Data.Workers{
		if temp2.ID == 1{
			continue//跳过admin账号
		}
		err = DownloadImage(temp2.Nums)
		if err!= nil{
			panic(err)
		}
	}
	//3. 读取所有图片
	file_engine,err := NewFaceEngine(DetectModeImage,OrientPriority0,1,EnableFaceDetect|EnableAge|EnableGender|EnableFaceRecognition)
	if err != nil{
		log.Fatal(err)
	}
	format := ".png"
	for _,v := range res2.Data.Workers{
		if v.ID == 1{
			continue//跳过admin账号
		}
		obj := Obj{}
		obj.id = int64(v.ID)
		obj.name = v.Name
		obj.nums = v.Nums
		obj.score = int64(v.Score)
		obj.Image = util.GetResizedImageInfo(localPath + obj.nums + format)
		obj.FaceInfos, err = file_engine.DetectFaces(obj.Image.Width, obj.Image.Height, ColorFormatBGR24, obj.Image.DataUInt8)
		if err != nil{
			log.Fatal("提取信息失败1",err)
		}
		if obj.FaceInfos.FaceDataInfoList == nil{
			log.Fatal("提取信息失败2",err)
		}
		obj.FaceInfo.DataInfo = obj.FaceInfos.FaceDataInfoList[0]
		obj.FaceInfo.FaceOrient = obj.FaceInfos.FaceOrient[0]
		obj.FaceInfo.FaceOrient = obj.FaceInfos.FaceOrient[0]
		obj.result,err = file_engine.FaceFeatureExtract(obj.Image.Width,obj.Image.Height,ColorFormatBGR24,obj.Image.DataUInt8,obj.FaceInfo,0,0)
		objs = append(objs,obj)
	}
}
// 激活SDK
func initFont() {
	fonts := g.Context.IO().Fonts()

	ranges := imgui.NewGlyphRanges()

	builder := imgui.NewFontGlyphRangesBuilder()
	builder.AddRanges(fonts.GlyphRangesChineseFull())
	builder.BuildRanges(ranges)

	fontPath := "c:/Alibaba-PuHuiTi-Light.ttf"
	fonts.AddFontFromFileTTFV(fontPath, 16, imgui.DefaultFontConfig, ranges.Data())
	title = fonts.AddFontFromFileTTFV(fontPath, 24, imgui.DefaultFontConfig, ranges.Data())
}

func loop(){
	if ok {
		now = strconv.FormatInt(time.Now().Unix(),10)//
	}
	ok = false
		g.SingleWindow("确认信息").IsOpen(&showWindow).Flags(g.WindowFlagsNone).Layout(
			g.Label("信息确认").Font(&title),
			//基本信息组
			g.Label("员工id:" + strconv.FormatInt(objs[maxindex].id,10)),
			g.Label("员工编号:" + objs[maxindex].nums),
			g.Label("员工姓名:" + objs[maxindex].name),
			g.Line(
				g.Label("员工证件照:"),
				g.ImageWithUrl(baseurl+ fmt.Sprintf("%s.png?time=%s",objs[maxindex].nums,now)),
			),
			g.Label("员工信誉分:" + strconv.FormatInt(objs[maxindex].score,10)),
			//按钮组
			g.Label("rfid标签").Font(&title),
			g.RangeBuilder("Labels",str, func(i int, v interface{}) g.Widget {
				return g.Label(v.(string))
			}),
			g.Line(
				g.Button("确认借出").Size(100,50).OnClick(Confirm),
			),

		)
}
func Confirm(){
	g.SingleWindow("message box").Layout(
		g.PrepareMsgbox(),
	)
	var err error
	client := &http.Client{}
	request,err := http.NewRequest("GET",fmt.Sprintf("%swork/workers", dataurl),nil)
	if err != nil{
		log.Fatal(err)
	}
	request.Header.Add("Authorization", token) //携带token访问
	temp,_ := client.Do(request)
	response,err := ioutil.ReadAll(temp.Body)
	if err != nil{
		log.Fatal(err)
	}
	err = json.Unmarshal(response, &AllRes)
	if err != nil{
		log.Println(err)
		g.Msgbox("error","reason:" + err.Error())
		return
	}
	if AllRes.Code != 200{
		log.Println(AllRes.Msg)
		g.Msgbox("error","reason:" + AllRes.Msg)
		return
	}
	fmt.Println(AllRes)
	//成功，请求出借
	for _,v := range devices.device{
		_,err := postBorrow(v.ID)
		if err != nil{
			g.Msgbox("error",err.Error())
		}
	}
}
func refresh(){
	//传入一个rfid
	g.Update()//更新界面
}
func test(callback func()){
	wnd = g.NewMasterWindow("租借信息确认",400,400,g.MasterWindowFlagsNotResizable, initFont)
	wnd.Run(loop)
	fmt.Println(123)
}
func main() {
	preResult = false
	//创建图形化界面
	var err error
	// 初始化人脸引擎
	engine, err = NewFaceEngine(DetectModeVideo,
		OrientPriority0,
		1,
		EnableFaceDetect|EnableAge|EnableGender|EnableFaceRecognition|EnableLiveness)
	if err != nil {
		panic(err)
	}
	GetFiles()
	media, err = gocv.VideoCaptureDevice(0) //根据id打开摄像头（我没有内置摄像头，所以是USB，惨惨兮兮）
	if err != nil {
		panic(err)
	}
	// 整个窗口方便看效果
	window = gocv.NewWindow("face detect")
	// 获取视频宽度
	w := media.Get(gocv.VideoCaptureFrameWidth)
	// 获取视频高度
	h := media.Get(gocv.VideoCaptureFrameHeight)
	// 调整窗口大小
	window.ResizeWindow(int(w), int(h))
	for{
		img := gocv.NewMat()
		media.Read(&img)
		if img.Empty() {
			continue
		}
		detectFace(engine, &img) //人脸识别
		window.IMShow(img)
		window.WaitKey(30)
		// 图片处理完毕记得关闭以释放内存
		img.Close()
	}
	// 收尾工作
	media.Close()
	engine.Destroy()
	window.Close()
}

// 虹软开始干活
func detectFace(engine *FaceEngine, img *gocv.Mat) bool{
	dataPtr, err := img.DataPtrUint8()//转换为ImageData所需类型
	if err != nil {
		fmt.Printf("%v\n", err)
		return false
	}
	imageData := ImageData{
		PixelArrayFormat: ColorFormatBGR24,
		Width:            img.Cols(),
		Height:           img.Rows(),
	}
	imageData.WidthStep[0] = img.Step()
	imageData.ImageData[0] = dataPtr
	faceInfo, err := engine.DetectFacesEx(imageData)//预处理
	if err != nil {
		fmt.Printf("%v\n", err)
		return false
	}
	if faceInfo.FaceNum > 0 {
		temp := SingleFaceInfo{
			FaceOrient: faceInfo.FaceOrient[0],
			FaceRect: faceInfo.FaceRect[0],
			DataInfo: faceInfo.FaceDataInfoList[0],
		}
		//对比一下
		temp1,err := engine.FaceFeatureExtractEx(imageData,temp,0,0)
		if err != nil{
			fmt.Print(1)
			log.Fatal(err)
		}
		var max float32 = 0.0
		i := 0
		for _,v := range objs {
			level,err := engine.FaceFeatureCompare(temp1,v.result)
			if err != nil{
				fmt.Print(2)
				log.Fatal(err)
			}
			if level > max {
				max = level
				maxindex = i
			}
			i++
		}
		err = engine.ProcessEx(imageData, faceInfo, EnableAge|EnableGender|EnableLiveness)
		for idx := 0; idx < int(faceInfo.FaceNum); idx++ {
			rect := image.Rect(int(faceInfo.FaceRect[idx].Left),
				int(faceInfo.FaceRect[idx].Top),
				int(faceInfo.FaceRect[idx].Right),
				int(faceInfo.FaceRect[idx].Bottom))
			// 把人脸框起来
			gocv.Rectangle(img, rect, color.RGBA{G: 255}, 2)
			if err == nil {
				age, _ := engine.GetAge()
				gender, _ := engine.GetGender()
				live,_ := engine.GetLivenessScore()
				var ageResult string
				var genderResult string
				if live.IsLive[idx] != 1 {
					//假体
					preResult = false
					showText := "prosthesis"
					gocv.PutText(img,fmt.Sprintf("%s",showText),
						image.Pt(int(faceInfo.FaceRect[idx].Right+2), int(faceInfo.FaceRect[idx].Top+10)),
						gocv.FontHersheyPlain,
						1,
						color.RGBA{R: 255},
						1,
					)
					return false
				}
				if age.AgeArray[idx] <= 0 {
					ageResult = "N/A"
				} else {
					ageResult = strconv.Itoa(int(age.AgeArray[idx]))
				}
				if gender.GenderArray[idx] < 0 {
					genderResult = "N/A"
				} else if gender.GenderArray[idx] == 0 {
					genderResult = "Male"
				} else {
					genderResult = "Female"
				}

				gocv.PutText(img,
					fmt.Sprintf("Age: %s", ageResult),
					image.Pt(int(faceInfo.FaceRect[idx].Right+2), int(faceInfo.FaceRect[idx].Top+10)),
					gocv.FontHersheyPlain,
					1,
					color.RGBA{R: 255},
					1)
				gocv.PutText(img,
					fmt.Sprintf("Gender: %s", genderResult),
					image.Pt(int(faceInfo.FaceRect[idx].Right+2), int(faceInfo.FaceRect[idx].Top+25)),
					gocv.FontHersheyPlain,
					1,
					color.RGBA{R: 255},
					1)
				if max > 0.8{
					//也只有到0.8以上的相似度我们才会允许员工借走
					gocv.PutText(img,
						fmt.Sprintf("nums: %s", objs[maxindex].nums),
						image.Pt(int(faceInfo.FaceRect[idx].Right+2), int(faceInfo.FaceRect[idx].Top+40)),
						gocv.FontHersheyPlain,
						1,
						color.RGBA{R: 255},
						1)
					gocv.PutText(img,
						fmt.Sprintf("simiar: %f", max),
						image.Pt(int(faceInfo.FaceRect[idx].Right+2), int(faceInfo.FaceRect[idx].Top+55)),
						gocv.FontHersheyPlain,
						1,
						color.RGBA{R: 255},
						1)
					if preResult == true{
						//判断如果感应到了rfid,读取rfid的租借信息。
						count++
						if count < 10{
							//性能优化，过多扫描rfid会对设备造成负担
							return true
						}
						count = 0
						if Rfid = RfidUtils.GetNearRfid();Rfid != nil {
							//每次扫描到要清空str
							getToken()//请求一次token
							str = nil
							devices.device = nil
							showWindow = true
							for _,v := range Rfid{
								var err error
								client := &http.Client{}
								request,err := http.NewRequest("GET",fmt.Sprintf("%sproduct/rfid/%s", dataurl,hex.EncodeToString(v)),nil)
								if err != nil{
									log.Fatal(err)
								}
								request.Header.Add("Authorization", token) //携带token访问
								temp,_ := client.Do(request)
								response,err := ioutil.ReadAll(temp.Body)
								if err != nil{
									log.Fatal(err)
								}
								fmt.Println(string(response))
								res3 := &res3{}
								err = json.Unmarshal(response, &res3)
								if err != nil{
									log.Println("读取信息失败")
								}
								str = append(str,res3.Name)
								devices.device = append(devices.device,*res3)
							}
							ok = true
							test(loop)
						}else{
							//提供开门操作，是员工
						}
						return true
					}else{
						//上一次检测是假体或置信度低于0.8，重新判断一次，同时gocv的puttext由于某种未知原因渲染的是上次结果，这也也可以保证渲染信息准确。
						preResult = true
					}
				}else{
					preResult = false
				}
			}
		}
	}
	return false
}