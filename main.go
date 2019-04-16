package main

import (
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"gopkg.in/urfave/cli.v1"
	"io/ioutil"
	"log"
	"os"
	"reflect"
)

var occuranceCounterMap int64 = 0
var occuranceCounterInterface int64 = 0
var occuranceCounterNil int64 = 0

/*
NOTE: Code adapted from https://gist.github.com/mcanevet/bf0a1a56dcc54ff320b877ce3fd31c4d
NOTE: adapted from https://play.golang.org/p/8jlJUbEJKf to support merging slices
*/

// merge merges the two JSON-marshalable values x1 and x2,
// preferring x1 over x2.
//
// It returns an error if x1 or x2 cannot be JSON-marshaled.
func merge(x1, x2 interface{}) (interface{}, error) {
	data1, err := json.Marshal(x1)
	if err != nil {
		return nil, err
	}
	data2, err := json.Marshal(x2)
	if err != nil {
		return nil, err
	}
	var j1 interface{}
	err = json.Unmarshal(data1, &j1)
	if err != nil {
		return nil, err
	}
	var j2 interface{}
	err = json.Unmarshal(data2, &j2)
	if err != nil {
		return nil, err
	}
	return merge1(j1, j2), nil
}

func merge1(x1, x2 interface{}) interface{} {
	switch x1 := x1.(type) {
	case map[string]interface{}:
		occuranceCounterMap += 1
		x2, ok := x2.(map[string]interface{})
		if !ok {
			return x1
		}
		for k, v2 := range x2 {
			if v1, ok := x1[k]; ok {
				x1[k] = merge1(v1, v2)
			} else {
				x1[k] = v2
			}
		}
	case []interface{}:
		occuranceCounterInterface += 1
		x2, ok := x2.([]interface{})
		//fmt.Println("Call Number: "+ strconv.Itoa(int(occuranceCounterInterface)))
		//fmt.Println(x1)
		//fmt.Println(x2)
		if !ok {
			return x1
		}
		for i := 0; i < len(x1) && i < len(x2); i++ {
			if reflect.DeepEqual(x1[i], x2[i]) {
				//todo: maybe use for verbose-information?
				//fmt.Println("On Call Number: "+ strconv.Itoa(int(occuranceCounterInterface)) + " The Elements are equal. The Elements where:")
				//fmt.Println(x1)
				//fmt.Println(x2)
				//fmt.Println("Returning x1")
				return x1
			}
		}
		for i := range x2 {
			x1 = append(x1, x2[i])
		}
		return x1
	case nil:
		occuranceCounterNil += 1
		// merge(nil, map[string]interface{...}) -> map[string]interface{...}
		x2, ok := x2.(map[string]interface{})
		if ok {
			return x2
		}
	}
	return x1
}

func startup(file1 string, file2 string, output string) error {
	var config interface{}

	files := []string{file1, file2}
	for i := range files {
		// Read
		raw, err := ioutil.ReadFile(files[i])
		if err != nil {
			fmt.Println(err.Error())
		}

		// Unmarshal
		var c map[string]interface{}
		err = yaml.Unmarshal(raw, &c)
		if err != nil {
			fmt.Println(err.Error())
		}

		// Merge
		config, err = merge(c, config)
		if err != nil {
			fmt.Println(err.Error())
		}
	}
	data, err := json.Marshal(config)
	if err != nil {
		fmt.Println(err.Error())
	}

	err = ioutil.WriteFile(output, data, 0644)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("Successfully merged files.")
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "mergeJSON"
	app.Usage = "merges two JSON files, removing doubles."
	app.Action = func(c *cli.Context) error {
		if c.NArg() < 3 {
			fmt.Println("Please use the following format: mergeJSON file1 file2 output")
			return nil
		}
		err := startup(c.Args().Get(0), c.Args().Get(1), c.Args().Get(2))
		if err != nil {
			log.Fatal(err)
		}
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
