package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	//testSelect()

	testNoRcvr()
	//var m map[string]int
	//m["shawn"] = 13

	//slc := make([]string, 5)
	//slc[0] = "shawn"

	//a := [5]string{"a", "b", "c", "d", "e"}
	//
	//s := a[1:2:3]
	//s2 := a[1:2]
	//s3 := a[1:3]
	//s4 := a[2:3]
	//fmt.Println(s)
	//fmt.Println(s2)
	//fmt.Println(s3)
	//fmt.Println(s4)

	// 声明一个切片，分配内存，但没有初始化
	//var slc []string
	//slc[0] = "lili" // 这会报错，因为没有初始化，不能进行赋值
	//slc = []string{"xxx"} // 初始化，并写入一个元素，cap只有1
	//slc[0] = "yyy"

	// 声明并初始化一个切片
	//slc2 := make([]string, 2)
	//slc2[2] = "xxx"

	// 声明一个数组，分配内存
	//var slc3 [5]string
	// 与切片不同，数组声明后就可以开始赋值
	//slc3[0] = "xxx"

	//fmt.Println(slc)
	//fmt.Println(slc2)
	//fmt.Println(slc3)
	//fmt.Printf("%p", slc)

	//t := &Test{}
	//
	//typ := reflect.TypeOf(t)
	//val := reflect.ValueOf(t)
	//
	//numMethod := typ.NumMethod()
	//for i := 0; i < numMethod; i++ {
	//	// 获取第一个方法
	//	m := typ.Method(i)
	//	mTyp := m.Type
	//	// 参数数量
	//	mIn := mTyp.NumIn()
	//
	//	vArr := make([]reflect.Value, mIn)
	//	vArr[0] = val // 第一个参数必须是实例本身
	//	for j := 1; j < mIn; j++ {
	//		in := mTyp.In(j)
	//		switch in.Kind() {
	//		case reflect.String:
	//			s := fmt.Sprintf("shawn%d", j)
	//			v := reflect.ValueOf(s)
	//			vArr[j] = v
	//			//vArr = append(vArr, v)
	//		case reflect.Bool:
	//			vArr[j] = reflect.ValueOf(false)
	//			//fmt.Println(v)
	//			//vArr = append(vArr, )
	//		}
	//	}
	//	f := m.Func
	//	rsp := f.Call(vArr)
	//	fmt.Println(rsp)
	//	for _, v := range rsp {
	//		switch v.Type().Name() {
	//		case "error":
	//			//tmp := (error)(v)
	//			//fmt.Println(v.())
	//		}
	//	}
	//}
}

func testSelect() {

	var chs [3]chan int

	for i := 0; i < 3; i++ {
		c := make(chan int)
		go func() {
			time.Sleep(time.Second * 3)

			select {
			case c <- i:
				fmt.Println("select success")
			default:
			}

			fmt.Println("sleep end:", i)
		}()
		chs[i] = c
	}

	fmt.Println(runtime.NumGoroutine())
	//select {
	//case num := <-chs[0]:
	//	fmt.Println("chan 0 get:", num)
	//case num := <-chs[1]:
	//	fmt.Println("chan 1 get:", num)
	//case num := <-chs[2]:
	//	fmt.Println("chan 2 get:", num)
	//	//default:
	//	//	fmt.Println("default")
	//}
	time.Sleep(time.Second * 5)
	fmt.Println(runtime.NumGoroutine())
	time.Sleep(time.Second * 5)
	fmt.Println(runtime.NumGoroutine())
	time.Sleep(time.Second * 5)
	fmt.Println(runtime.NumGoroutine())
}

func testNoRcvr() {
	ch1 := make(chan bool)
	ch2 := make(chan bool)

	go func() {
		time.Sleep(time.Second * 3)
		ch1 <- true
	}()

	select {
	case <-ch1:
		fmt.Println("receive from ch1")
	case <-ch2:
		fmt.Println("receive from ch2")
	default:
		fmt.Println("default")
	}
}
