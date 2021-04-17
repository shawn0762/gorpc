package main

import "fmt"

func main() {


	m := p()
	for k, v := range m {
		fmt.Printf("key = %s, val = %s \n", k, v)
	}

	//u1 := user{
	//	name: "lili",
	//	age:  18,
	//}
	//
	////jsonStr := "{\"name\":\"Shawn\",\"age\":18}"
	//jsonStr, err := json.Marshal(u1)
	//if err != nil {
	//	fmt.Println("json marshal wrong")
	//	return
	//}
	//
	//fmt.Println(jsonStr)
	//buf := bytes.NewBuffer(jsonStr)
	//
	//r := io.Reader(buf)
	//decoder := json.NewDecoder(r)
	//
	//var u user
	//decoder.Decode(&u)
	//
	//fmt.Println("JSON 格式数据：%s", u)
}

type user struct {
	name string
	age  int
}

func p() map[string]*user {
	stau := []user{
		{name: "shawn", age: 18},
		{name: "Lili", age: 21},
	}
	m := make(map[string]*user)
	for _, stu := range stau{
		//fmt.Println("name: ", stu.name)
		//fmt.Println("age: ", stu.age)
		m[stu.name] = &stu
	}
	return m
}
