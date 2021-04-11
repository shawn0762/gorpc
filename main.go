package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

func main() {

	u1 := user{
		name: "lili",
		age:  18,
	}

	//jsonStr := "{\"name\":\"Shawn\",\"age\":18}"
	jsonStr, err := json.Marshal(u1)
	if err != nil {
		fmt.Println("json marshal wrong")
		return
	}

	fmt.Println(jsonStr)
	buf := bytes.NewBuffer(jsonStr)

	r := io.Reader(buf)
	decoder := json.NewDecoder(r)

	var u user
	decoder.Decode(&u)

	fmt.Println("JSON 格式数据：%s", u)
}

type user struct {
	name string
	age  int
}
