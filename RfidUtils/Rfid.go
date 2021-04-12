package RfidUtils
import (
	"crypto/md5"
	"errors"
	serial "github.com/tarm/goserial"
	"io"
	"log"
)
var (
	s       io.ReadWriteCloser
	cardNum [][]byte
)
func init(){
	var err error
	cfg := &serial.Config{Name: "COM6", Baud: 115200, ReadTimeout: 50 /*毫秒*/}
	s,err = serial.OpenPort(cfg)
	FatalErr(err)
}
func GetRes(req []byte)[]byte{
	//获取有效位函数
	i:=0
	for ;i<128;i++{
		if req[i] == 126 && req[i+1] ==0 {
			//读到终止符号
			break
		}
	}
	return req[:i+1]
}
func GetReadEpc(req []byte)[]byte{
	//获取epc区域，做地址偏移，找到对应的数据区域
	return req[8:20]
}
func CheckSum(req []byte)int{
	//计算sum值
	sum := 0
	for i:=1;i<len(req);i++{
		sum += int(req[i])
	}
	return sum % 256
}
func AddLastByte(req []byte)[]byte{
	req = append(req,byte(CheckSum(req)),byte(0x7E))
	return req
}
func FatalErr(err error){
	if err != nil{
		log.Fatal(err)
	}
}
func appendSlice(ele1 []byte,ele2 []byte)[]byte{
	for i:=0;i<len(ele2);i++{
		ele1 = append(ele1,ele2[i])
	}
	return ele1
}
func GetIoReaderData()(int,[]byte){
	//只读取一条返回指令哦
	data := make([]byte,128)
	for i:=0;i<500;i++{
		if s == nil{
			FatalErr(errors.New("无法读取s"))
		}
		n,err := s.Read(data)
		FatalErr(err)
		if n!=0{
			return n,GetRes(data)//只返回有效区域
		}
	}
	return 0,nil
}
func Success(data []byte)bool{
	temp := []byte{0xBB,0x01,0x0C,0x00,0x01,0x00,0x0E,0x7E}
	for k,_ := range temp{
		if temp[k] != data[k]{
			return false
		}
	}
	return true
}
func Select(buf []byte,needEpc bool)[]byte{
	//传入要选择的卡片,返回响应结果
	// BB 00 0C 00 13 23 00 00 00 00 60 00
	command := []byte{0xBB,0x00,0x0C,0x00,0x13,0x23,0x00,0x00,0x00,0x00,0x60,0x00}
	if needEpc{
		command = appendSlice(command,GetReadEpc(buf))
	}else{
		command = appendSlice(command,buf)
	}
	return Command(command)
}
func Write()[]byte{
	//封装的Write函数
	/*
		BB 00 49 00 11 00 00 00 00 03 00 00 00 04 01 02 03 04 05 06 07 08 85 7E
	*/
	/*command := []byte{0xBB,0x00,0x49,0x00,0x11,0x00,0x00,0x00,0x00,0x01,
	0x00,0x00,0x00,0x04,0x01,0x02,0x03,0x04,0x05,0x06,0x07,0x08}*/
	/*帧类型
	Type: 0x00 指令代码 Command: 0x49 指令参数长度 PL: 0x000D
	Access Password: 0x0000FFFF 标签数据存储区 MemBank: 0x03 标签数据区地址偏移
	SA: 0x0000 数据长度 DL: 0x0002 写入数据 DT: 0x12345678 校验位 Checksum: 0x6D*/
	command := []byte{0xBB,0x00,0x49,0x00,0x11,0x00,0x00,0x00,0x00,0x01,
		0x00,0x00,0x00,0x04,0x01,0x02,0x03,0x04,0x05,0x06,0x07,0x08}
	return Command(command)
}
func WriteErr(req []byte)error{
	if req == nil || len(req) < 6{
		return errors.New("传入的数据不正确")
	}
	if req[0] == 0xbb && req[1] == 0x01 && req[2] == 0x49{
		return nil
	}
	if req[0] == 0xbb && req[1] == 0x01 && req[2] == 0xff{
		if req[5] == 0x10{
			return errors.New("没有找到指定卡号")
		}
		if req[5] == 0x10{
			return errors.New("访问密码错误")
		}
		if req[5] == 0xb3{
			return errors.New("超出读写范围")
		}
		return errors.New("未定义的错误")
	}
	return errors.New("格式不正确，请检查传入的数据是否为写入响应结果")
}
func StartReadAll(){
	//开始群读
	command := []byte{0xBB,0x00,0x27,0x00,0x03,0x02,0xff,0xff}
	command = AddLastByte(command)
	_,err := s.Write(command)
	FatalErr(err)
}
func StopReadAll(){
	//用于用户主动放弃群读操作
	command := []byte{0xBB,0x00,0x28,0x00,0x00}
	command = AddLastByte(command)
	_,err := s.Write(command)
	FatalErr(err)
}
func AddToArray(temp []byte){
	//作用：将cardNum去重,md5直接对比，撞得概率贼小，中了就和空难一样稀奇
	ok := true
	for _,v := range cardNum{
		if md5.Sum(v) == md5.Sum(temp){
			ok = false
			break
		}
	}
	if ok{
		cardNum = append(cardNum,temp)
	}
}
func Command(command []byte)[]byte{
	//对执行Command一个封装
	command = AddLastByte(command)
	_,err := s.Write(command)
	FatalErr(err)
	_,temp := GetIoReaderData()
	return temp
}
func Empty(data []byte)bool{
	for _,v := range data{
		if v!= 0{
			return false
		}
	}
	return  true
}
func GetNearRfid()[][]byte{
	cardNum = nil
	StartReadAll()
	for i := 0;i<200;i++{
		buf := make([]byte,128)
		n, _ := s.Read(buf)
		if n >0 {
			if !Empty(GetReadEpc(buf)) {
				AddToArray(GetReadEpc(buf))
			}
		}
	}
	StopReadAll()//关闭执行
	return cardNum
}