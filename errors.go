// Goro
//
// Created by Yakka
// http://theyakka.com
//
// Copyright (c) 2019 Yakka LLC.
// All rights reserved.
// See the LICENSE file for licensing details and requirements.

package goro

type ErrorInfoMap map[string]interface{}

type RoutingError struct {
	StatusCode int
	ErrorCode  RouterErrorCode
	Error      error
	Message    string
	Info       ErrorInfoMap
}

func EmptyRoutingError() RoutingError {
	return RoutingError{
		StatusCode: 0,
		ErrorCode:  0,
		Error:      nil,
		Message:    "",
		Info:       nil,
	}
}

func IsEmptyRoutingError(re RoutingError) bool {
	return re.StatusCode == 0 && re.Message == "" && re.Info == nil &&
		re.ErrorCode == 0 && re.Error == nil
}
