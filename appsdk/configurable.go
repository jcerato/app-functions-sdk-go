//
// Copyright (c) 2019 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package appsdk

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/jcerato/app-functions-sdk-go/appcontext"
	"github.com/jcerato/app-functions-sdk-go/pkg/transforms"
	"github.com/jcerato/app-functions-sdk-go/pkg/util"
)

const (
	ValueDescriptors    = "valuedescriptors"
	DeviceNames         = "devicenames"
	FilterOut           = "filterout"
	Key                 = "key"
	InitVector          = "initvector"
	Url                 = "url"
	MimeType            = "mimetype"
	PersistOnError      = "persistonerror"
	Cert                = "cert"
	SkipVerify          = "skipverify"
	Qos                 = "qos"
	Retain              = "retain"
	AutoReconnect       = "autoreconnect"
	DeviceName          = "devicename"
	ReadingName         = "readingname"
	Rule                = "rule"
	BatchThreshold      = "batchthreshold"
	TimeInterval        = "timeinterval"
	SecretHeaderName1   = "secretheadername1"
	SecretHeaderName2   = "secretheadername2"
	SecretPath          = "secretpath"
	BrokerAddress       = "brokeraddress"
	ClientID            = "clientid"
	Topic               = "topic"
	AuthMode            = "authmode"
	Tags                = "tags"
	ResponseContentType = "responsecontenttype"
)

// AppFunctionsSDKConfigurable contains the helper functions that return the function pointers for building the configurable function pipeline.
// They transform the parameters map from the Pipeline configuration in to the actual actual parameters required by the function.
type AppFunctionsSDKConfigurable struct {
	Sdk *AppFunctionsSDK
}

// FilterByDeviceName - Specify the devices of interest to  filter for data coming from certain sensors.
// The Filter by Device transform looks at the Event in the message and looks at the devices of interest list,
// provided by this function, and filters out those messages whose Event is for devices not on the
// devices of interest.
// This function will return an error and stop the pipeline if a non-edgex
// event is received or if no data is recieved.
// For example, data generated by a motor does not get passed to functions only interested in data from a thermostat.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) FilterByDeviceName(parameters map[string]string) appcontext.AppFunction {
	deviceNames, ok := parameters[DeviceNames]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + DeviceNames)
		return nil
	}

	filterOutBool := false
	filterOut, ok := parameters[FilterOut]
	if ok {
		var err error
		filterOutBool, err = strconv.ParseBool(filterOut)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error("Could not convert filterOut value to bool " + filterOut)
			return nil
		}
	}

	deviceNamesCleaned := util.DeleteEmptyAndTrim(strings.FieldsFunc(deviceNames, util.SplitComma))
	transform := transforms.Filter{
		FilterValues: deviceNamesCleaned,
		FilterOut:    filterOutBool,
	}
	dynamic.Sdk.LoggingClient.Debug("Device Name Filters", DeviceNames, strings.Join(deviceNamesCleaned, ","))

	return transform.FilterByDeviceName
}

// FilterByValueDescriptor - Specify the value descriptors of interest to filter for data from certain types of IoT objects,
// such as temperatures, motion, and so forth, that may come from an array of sensors or devices. The Filter by Value Descriptor assesses
// the data in each Event and Reading, and removes readings that have a value descriptor that is not in the list of
// value descriptors of interest for the application.
// This function will return an error and stop the pipeline if a non-edgex
// event is received or if no data is recieved.
// For example, pressure reading data does not go to functions only interested in motion data.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) FilterByValueDescriptor(parameters map[string]string) appcontext.AppFunction {
	valueDescriptors, ok := parameters[ValueDescriptors]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + ValueDescriptors)
		return nil
	}

	filterOutBool := false
	filterOut, ok := parameters[FilterOut]
	if ok {
		var err error
		filterOutBool, err = strconv.ParseBool(filterOut)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error("Could not convert filterOut value to bool " + filterOut)
			return nil
		}
	}

	valueDescriptorsCleaned := util.DeleteEmptyAndTrim(strings.FieldsFunc(valueDescriptors, util.SplitComma))
	transform := transforms.Filter{
		FilterValues: valueDescriptorsCleaned,
		FilterOut:    filterOutBool,
	}
	dynamic.Sdk.LoggingClient.Debug("Value Descriptors Filter", ValueDescriptors, strings.Join(valueDescriptorsCleaned, ","))

	return transform.FilterByValueDescriptor
}

// TransformToXML transforms an EdgeX event to XML.
// It will return an error and stop the pipeline if a non-edgex
// event is received or if no data is recieved.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) TransformToXML() appcontext.AppFunction {
	transform := transforms.Conversion{}
	return transform.TransformToXML
}

// TransformToJSON transforms an EdgeX event to JSON.
// It will return an error and stop the pipeline if a non-edgex
// event is received or if no data is recieved.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) TransformToJSON() appcontext.AppFunction {
	transform := transforms.Conversion{}
	return transform.TransformToJSON
}

// MarkAsPushed will make a request to CoreData to mark the event that triggered the pipeline as pushed.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) MarkAsPushed() appcontext.AppFunction {
	transform := transforms.CoreData{}
	return transform.MarkAsPushed
}

// PushToCore pushes the provided value as an event to CoreData using the device name and reading name that have been set. If validation is turned on in
// CoreServices then your deviceName and readingName must exist in the CoreMetadata and be properly registered in EdgeX.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) PushToCore(parameters map[string]string) appcontext.AppFunction {
	deviceName, ok := parameters[DeviceName]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + DeviceName)
		return nil
	}
	readingName, ok := parameters[ReadingName]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + readingName)
		return nil
	}
	deviceName = strings.TrimSpace(deviceName)
	readingName = strings.TrimSpace(readingName)
	dynamic.Sdk.LoggingClient.Debug("PushToCore Parameters", DeviceName, deviceName, ReadingName, readingName)
	transform := transforms.CoreData{
		DeviceName:  deviceName,
		ReadingName: readingName,
	}
	return transform.PushToCoreData
}

// CompressWithGZIP compresses data received as either a string,[]byte, or json.Marshaler using gzip algorithm and returns a base64 encoded string as a []byte.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) CompressWithGZIP() appcontext.AppFunction {
	transform := transforms.Compression{}
	return transform.CompressWithGZIP
}

// CompressWithZLIB compresses data received as either a string,[]byte, or json.Marshaler using zlib algorithm and returns a base64 encoded string as a []byte.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) CompressWithZLIB() appcontext.AppFunction {
	transform := transforms.Compression{}
	return transform.CompressWithZLIB
}

// EncryptWithAES encrypts either a string, []byte, or json.Marshaller type using AES encryption.
// It will return a byte[] of the encrypted data.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) EncryptWithAES(parameters map[string]string) appcontext.AppFunction {
	key, ok := parameters[Key]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + Key)
		return nil
	}
	initVector, ok := parameters[InitVector]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + InitVector)
		return nil
	}
	transforms := transforms.Encryption{
		Key:                  key,
		InitializationVector: initVector,
	}
	return transforms.EncryptWithAES
}

// HTTPPost will send data from the previous function to the specified Endpoint via http POST. If no previous function exists,
// then the event that triggered the pipeline will be used. Passing an empty string to the mimetype
// method will default to application/json.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) HTTPPost(parameters map[string]string) appcontext.AppFunction {
	var err error

	url, ok := parameters[Url]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("HTTPPost Could not find " + Url)
		return nil
	}
	mimeType, ok := parameters[MimeType]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("HTTPPost Could not find " + MimeType)
		return nil
	}

	// PersistOnError is optional and is false by default.
	persistOnError := false
	value, ok := parameters[PersistOnError]
	if ok {
		persistOnError, err = strconv.ParseBool(value)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("HTTPPost Could not parse '%s' to a bool for '%s' parameter", value, PersistOnError), "error", err)
			return nil
		}
	}

	url = strings.TrimSpace(url)
	mimeType = strings.TrimSpace(mimeType)

	secretHeaderName1 := parameters[SecretHeaderName1]
	secretHeaderName2 := parameters[SecretHeaderName2]
	secretPath := parameters[SecretPath]
	var transform transforms.HTTPSender
	if secretHeaderName1 != "" && secretPath != "" {
		transform = transforms.NewHTTPSenderWithSecretHeader(url, mimeType, persistOnError, secretHeaderName1, secretHeaderName2, secretPath)
	} else {
		transform = transforms.NewHTTPSender(url, mimeType, persistOnError)
	}
	dynamic.Sdk.LoggingClient.Debug("HTTPPost Parameters", Url, transform.URL, MimeType, transform.MimeType)
	return transform.HTTPPost
}

// HTTPPostJSON sends data from the previous function to the specified Endpoint via http POST with a mime type of application/json.
// If no previous function exists, then the event that triggered the pipeline will be used.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) HTTPPostJSON(parameters map[string]string) appcontext.AppFunction {
	parameters[MimeType] = "application/json"
	return dynamic.HTTPPost(parameters)
}

// HTTPPostXML sends data from the previous function to the specified Endpoint via http POST with a mime type of application/xml.
// If no previous function exists, then the event that triggered the pipeline will be used.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) HTTPPostXML(parameters map[string]string) appcontext.AppFunction {
	parameters[MimeType] = "application/xml"
	return dynamic.HTTPPost(parameters)
}

// HTTPPut will send data from the previous function to the specified Endpoint via http PUT. If no previous function exists,
// then the event that triggered the pipeline will be used. Passing an empty string to the mimetype
// method will default to application/json.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) HTTPPut(parameters map[string]string) appcontext.AppFunction {
	var err error

	url, ok := parameters[Url]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("HTTPPut Could not find " + Url)
		return nil
	}
	mimeType, ok := parameters[MimeType]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("HTTPPut Could not find " + MimeType)
		return nil
	}

	// PersistOnError is optional and is false by default.
	persistOnError := false
	value, ok := parameters[PersistOnError]
	if ok {
		persistOnError, err = strconv.ParseBool(value)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("HTTPPut Could not parse '%s' to a bool for '%s' parameter", value, PersistOnError), "error", err)
			return nil
		}
	}

	url = strings.TrimSpace(url)
	mimeType = strings.TrimSpace(mimeType)

	secretHeaderName1 := parameters[SecretHeaderName1]
	secretHeaderName2 := parameters[SecretHeaderName2]
	secretPath := parameters[SecretPath]
	var transform transforms.HTTPSender
	if secretHeaderName1 != "" && secretPath != "" {
		transform = transforms.NewHTTPSenderWithSecretHeader(url, mimeType, persistOnError, secretHeaderName1, secretHeaderName2, secretPath)
	} else {
		transform = transforms.NewHTTPSender(url, mimeType, persistOnError)
	}
	dynamic.Sdk.LoggingClient.Debug("HTTPPut Parameters", Url, transform.URL, MimeType, transform.MimeType)
	return transform.HTTPPut
}

// HTTPPutJSON sends data from the previous function to the specified Endpoint via http PUT with a mime type of application/json.
// If no previous function exists, then the event that triggered the pipeline will be used.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) HTTPPutJSON(parameters map[string]string) appcontext.AppFunction {
	parameters[MimeType] = "application/json"
	return dynamic.HTTPPut(parameters)
}

// HTTPPutXML sends data from the previous function to the specified Endpoint via http PUT with a mime type of application/xml.
// If no previous function exists, then the event that triggered the pipeline will be used.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) HTTPPutXML(parameters map[string]string) appcontext.AppFunction {
	parameters[MimeType] = "application/xml"
	return dynamic.HTTPPut(parameters)
}

// MQTTSend sends data from the previous function to the specified MQTT broker.
// If no previous function exists, then the event that triggered the pipeline will be used.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) MQTTSend(parameters map[string]string, addr models.Addressable) appcontext.AppFunction {
	var err error
	qos := 0
	retain := false
	autoReconnect := false
	// optional string params
	cert := parameters[Cert]
	key := parameters[Key]
	skipVerify := parameters[SkipVerify]

	qosVal, ok := parameters[Qos]
	if ok {
		qos, err = strconv.Atoi(qosVal)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error("Unable to parse " + Qos + " value")
			return nil
		}
	}
	retainVal, ok := parameters[Retain]
	if ok {
		retain, err = strconv.ParseBool(retainVal)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error("Unable to parse " + Retain + " value")
			return nil
		}
	}
	autoreconnectVal, ok := parameters[AutoReconnect]
	if ok {
		autoReconnect, err = strconv.ParseBool(autoreconnectVal)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error("Unable to parse " + AutoReconnect + " value")
			return nil
		}
	}
	dynamic.Sdk.LoggingClient.Debug("MQTT Send Parameters", "Address", addr, Qos, qosVal, Retain, retainVal, AutoReconnect, autoreconnectVal, Cert, cert, Key, key)

	var pair *transforms.KeyCertPair

	if len(cert) > 0 && len(key) > 0 {
		pair = &transforms.KeyCertPair{
			CertFile: cert,
			KeyFile:  key,
		}
	}

	// PersistOnError os optional and is false by default.
	persistOnError := false
	value, ok := parameters[PersistOnError]
	if ok {
		persistOnError, err = strconv.ParseBool(value)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Could not parse '%s' to a bool for '%s' parameter", value, PersistOnError), "error", err)
			return nil
		}
	}

	mqttConfig := transforms.MqttConfig{}
	mqttConfig.Qos = byte(qos)
	mqttConfig.Retain = retain
	mqttConfig.AutoReconnect = autoReconnect

	if skipVerify != "" {
		skipCertVerify, err := strconv.ParseBool(skipVerify)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Could not parse '%s' to a bool for '%s' parameter", skipVerify, SkipVerify), "error", err)
			return nil
		}

		mqttConfig.SkipCertVerify = skipCertVerify
	}

	sender := transforms.NewMQTTSender(dynamic.Sdk.LoggingClient, addr, pair, mqttConfig, persistOnError)
	return sender.MQTTSend
}

// SetOutputData sets the output data to that passed in from the previous function.
// It will return an error and stop the pipeline if data passed in is not of type []byte, string or json.Mashaler
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) SetOutputData(parameters map[string]string) appcontext.AppFunction {
	transform := transforms.OutputData{}

	value, ok := parameters[ResponseContentType]
	if ok && len(value) > 0 {
		transform.ResponseContentType = value
	}

	return transform.SetOutputData
}

// BatchByCount ...
func (dynamic AppFunctionsSDKConfigurable) BatchByCount(parameters map[string]string) appcontext.AppFunction {
	batchThreshold, ok := parameters[BatchThreshold]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + BatchThreshold)
		return nil
	}

	thresholdValue, err := strconv.Atoi(batchThreshold)
	if err != nil {
		dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Could not parse '%s' to an int for '%s' parameter", batchThreshold, BatchThreshold), "error", err)
		return nil
	}
	transform, err := transforms.NewBatchByCount(thresholdValue)
	if err != nil {
		dynamic.Sdk.LoggingClient.Error(err.Error())
	}
	dynamic.Sdk.LoggingClient.Debug("Batch by count Parameters", BatchThreshold, batchThreshold)
	return transform.Batch
}

// BatchByTime ...
func (dynamic AppFunctionsSDKConfigurable) BatchByTime(parameters map[string]string) appcontext.AppFunction {
	timeInterval, ok := parameters[TimeInterval]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + TimeInterval)
		return nil
	}
	transform, err := transforms.NewBatchByTime(timeInterval)
	if err != nil {
		dynamic.Sdk.LoggingClient.Error(err.Error())
	}
	dynamic.Sdk.LoggingClient.Debug("Batch by time Parameters", TimeInterval, timeInterval)
	return transform.Batch
}

// BatchByTimeAndCount ...
func (dynamic AppFunctionsSDKConfigurable) BatchByTimeAndCount(parameters map[string]string) appcontext.AppFunction {
	timeInterval, ok := parameters[TimeInterval]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + TimeInterval)
		return nil
	}
	batchThreshold, ok := parameters[BatchThreshold]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + BatchThreshold)
		return nil
	}
	thresholdValue, err := strconv.Atoi(batchThreshold)
	if err != nil {
		dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Could not parse '%s' to an int for '%s' parameter", batchThreshold, BatchThreshold), "error", err)
	}
	transform, err := transforms.NewBatchByTimeAndCount(timeInterval, thresholdValue)
	if err != nil {
		dynamic.Sdk.LoggingClient.Error(err.Error())
	}
	dynamic.Sdk.LoggingClient.Debug("Batch by time and count Parameters", BatchThreshold, batchThreshold, TimeInterval, timeInterval)
	return transform.Batch
}

// JSONLogic ...
func (dynamic AppFunctionsSDKConfigurable) JSONLogic(parameters map[string]string) appcontext.AppFunction {
	rule, ok := parameters[Rule]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + Rule)
		return nil
	}
	transform := transforms.NewJSONLogic(rule)
	return transform.Evaluate
}

// MQTTSecretSend
func (dynamic AppFunctionsSDKConfigurable) MQTTSecretSend(parameters map[string]string) appcontext.AppFunction {
	var err error
	qos := 0
	retain := false
	autoReconnect := false
	skipCertVerify := false

	brokerAddress, ok := parameters[BrokerAddress]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + BrokerAddress)
		return nil
	}
	topic, ok := parameters[Topic]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + Topic)
		return nil
	}

	secretPath, ok := parameters[SecretPath]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + SecretPath)
		return nil
	}
	authMode, ok := parameters[AuthMode]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + AuthMode)
		return nil
	}
	clientID, ok := parameters[ClientID]
	if !ok {
		dynamic.Sdk.LoggingClient.Error("Could not find " + ClientID)
		return nil
	}
	qosVal, ok := parameters[Qos]
	if ok {
		qos, err = strconv.Atoi(qosVal)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error("Unable to parse " + Qos + " value")
			return nil
		}
	}
	retainVal, ok := parameters[Retain]
	if ok {
		retain, err = strconv.ParseBool(retainVal)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error("Unable to parse " + Retain + " value")
			return nil
		}
	}
	autoreconnectVal, ok := parameters[AutoReconnect]
	if ok {
		autoReconnect, err = strconv.ParseBool(autoreconnectVal)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error("Unable to parse " + AutoReconnect + " value")
			return nil
		}
	}
	skipVerifyVal, ok := parameters[SkipVerify]
	if ok {
		skipCertVerify, err = strconv.ParseBool(skipVerifyVal)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Could not parse '%s' to a bool for '%s' parameter", skipVerifyVal, SkipVerify), "error", err)
			return nil
		}
	}
	mqttConfig := transforms.MQTTSecretConfig{
		Retain:         retain,
		SkipCertVerify: skipCertVerify,
		AutoReconnect:  autoReconnect,
		QoS:            byte(qos),
		BrokerAddress:  brokerAddress,
		ClientId:       clientID,
		SecretPath:     secretPath,
		Topic:          topic,
		AuthMode:       authMode,
	}
	// PersistOnError is optional and is false by default.
	persistOnError := false
	value, ok := parameters[PersistOnError]
	if ok {
		persistOnError, err = strconv.ParseBool(value)
		if err != nil {
			dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Could not parse '%s' to a bool for '%s' parameter", value, PersistOnError), "error", err)
			return nil
		}
	}
	transform := transforms.NewMQTTSecretSender(mqttConfig, persistOnError)
	return transform.MQTTSend
}

// AddTags adds the configured list of tags to Events passed to the transform.
// This function is a configuration function and returns a function pointer.
func (dynamic AppFunctionsSDKConfigurable) AddTags(parameters map[string]string) appcontext.AppFunction {
	tagsSpec, ok := parameters[Tags]
	if !ok {
		dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Could not find '%s' parameter", Tags))
		return nil
	}

	tagKeyValues := util.DeleteEmptyAndTrim(strings.FieldsFunc(tagsSpec, util.SplitComma))

	tags := make(map[string]string)
	for _, tag := range tagKeyValues {
		keyValue := util.DeleteEmptyAndTrim(strings.FieldsFunc(tag, util.SplitColon))
		if len(keyValue) != 2 {
			dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Bad Tags specification format. Expect comma separated list of 'key:value'. Got `%s`", tagsSpec))
			return nil
		}

		if len(keyValue[0]) == 0 {
			dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Tag key missing. Got '%s'", tag))
			return nil
		}
		if len(keyValue[1]) == 0 {
			dynamic.Sdk.LoggingClient.Error(fmt.Sprintf("Tag value missing. Got '%s'", tag))
			return nil
		}

		tags[keyValue[0]] = keyValue[1]
	}

	transform := transforms.NewTags(tags)
	dynamic.Sdk.LoggingClient.Debug("Add Tags", Tags, fmt.Sprintf("%v", tags))

	return transform.AddTags
}
