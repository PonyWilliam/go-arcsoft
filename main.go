package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	serial "github.com/tarm/goserial"
	"io"
	"log"
)
var s io.ReadWriteCloser
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
func Select(buf []byte)[]byte{
	//传入要选择的卡片,返回响应结果
	command := []byte{0xBB,0x00,0x0C,0x00,0x13,0x01,0x00,0x00,0x00,0x20,0x60,0x00}
	command = appendSlice(command,GetReadEpc(buf))
	command = AddLastByte(command)
	_,err := s.Write(command)
	FatalErr(err)
	_, temp := GetIoReaderData()
	return temp
}
func Write()[]byte{
	//封装的Write函数
	/*
		BB 00 49 00 11 00 00 00 00 03 00 00 00 04 01 02 03 04 05 06 07 08 85 7E
	*/
	command := []byte{0xBB,0x00,0x49,0x00,0x11,0x00,0x00,0x00,0x00,0x01,
		0x00,0x00,0x00,0x04,0x01,0x02,0x03,0x04,0x05,0x06,0x07,0x08}
	command = AddLastByte(command)
	fmt.Println(hex.EncodeToString(command))
	_,err := s.Write(command)
	FatalErr(err)
	_, temp := GetIoReaderData()
	return temp
}
func WriteErr(req []byte)error{
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
func main() {
	var err error
	cfg := &serial.Config{Name: "COM6", Baud: 115200, ReadTimeout: 50 /*毫秒*/}
	s,err = serial.OpenPort(cfg)
	FatalErr(err)
	command := []byte{0xBB,0x00,0x22,0x00,0x00}
	command = AddLastByte(command)
	n,err := s.Write(command)
	FatalErr(err)
	buf := make([]byte,128)
	for i:=0;i<500;i++{
		n,err = s.Read(buf)
		FatalErr(err)
		if n > 0{
			//获取数据
			if n == 8{
				fmt.Println("读取完毕")
				return
			}
			strTemp := hex.EncodeToString(GetReadEpc(buf))
			fmt.Println(strTemp)
			//command := []byte{0xBB,0x00,0x0C,0x00,0x07,0x23,0x00,0x00,0x00,0x00,0x60,0x00}
			temp := Select(buf)//封装的选择函数
			if Success(temp){
				res := Write()//封装的Write，先select再write
				fmt.Println(hex.EncodeToString(res))
				if err :=WriteErr(res);err!=nil {
					log.Fatal(err)
				}else{
					fmt.Println("写入成功")
				}
			}else{
				log.Fatal("好像出错了。。。")
			}
		}
	}
}