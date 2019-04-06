// Goro
//
// Created by Yakka
// http://theyakka.com
//
// Copyright (c) 2019 Yakka LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro

type Parameters struct {
	paramsMap map[string][]interface{}
}

func NewParametersWithMap(paramsMap map[string][]string) *Parameters {
	interfaceMap := stringMapToInterfaceMap(paramsMap)
	return &Parameters{
		paramsMap: interfaceMap,
	}
}

func (p *Parameters) Get(key string) []interface{} {
	return p.paramsMap[key]
}

func (p *Parameters) GetStrings(key string) []string {
	val := p.paramsMap[key]
	if val == nil || len(val) == 0 {
		return []string{}
	}
	outArr := make([]string, len(val))
	for i, v := range val {
		if n, ok := v.(string); ok {
			outArr[i] = n
		}
	}
	return outArr
}

func (p *Parameters) GetFirstString(key string) string {
	strArr := p.GetStrings(key)
	if len(strArr) == 0 {
		return ""
	}
	return strArr[0]
}

func (p *Parameters) GetInts(key string) []int {
	val := p.paramsMap[key]
	if val == nil || len(val) == 0 {
		return []int{}
	}
	outArr := make([]int, len(val))
	for i, v := range val {
		if n, ok := v.(int); ok {
			outArr[i] = n
		}
	}
	return outArr
}

func (p *Parameters) GetInt(key string) int {
	intArr := p.GetInts(key)
	if len(intArr) == 0 {
		return 0
	}
	return intArr[0]

}

// converts a map of lists of strings to a generic interface
func stringMapToInterfaceMap(stringMap map[string][]string) map[string][]interface{} {
	outMap := map[string][]interface{}{}
	for k, v := range stringMap {
		var outArr []interface{}
		for _, vi := range v {
			outArr = append(outArr, vi)
		}
		outMap[k] = outArr
	}
	return outMap
}
