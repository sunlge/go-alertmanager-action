package main

import (
	"alertmanager/pkg"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Dintalk接收信息
type Message struct {
	MsgType string `json:"msgtype"`
	Text struct {
		Content string `json:"content"`
		Mentioned_list string `json:"mentioned_list"`
		Mentioned_mobile_list string `json:"mentioned_mobile_list"`
	} `json:"text"`

}

type Alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:annotations`
	StartsAt    time.Time         `json:"startsAt"`
	EndsAt      time.Time         `json:"endsAt"`
}

// 定义一个接收alertmanager告警消息的结构体
type Notification struct {
	Version           string            `json:"version"`		  // api版本
	GroupKey          string            `json:"groupKey"`		  // 定义的告警规则的标签 {}/{action="delete",namespace="kube-plugin"}:{namespace="kube-plugin"}
	Status            string            `json:"status"`			  // 告警状态 firing
	Receiver          string            `json:receiver`			  // Receiver接收器
	GroupLabels       map[string]string `json:groupLabels`		  // ---
	CommonLabels      map[string]string `json:commonLabels`		  // 触发告警rule的标签 这里拿到的信息可能和实际告警的应用不匹配。
	CommonAnnotations map[string]string `json:commonAnnotations`  // value不一致的情况下无法获取，可以从Alert里面拿。
	ExternalURL       string            `json:externalURL`		  // ---
	Alerts            []Alert           `json:alerts`
}

type AlertSizeInfo struct {
	PodList 		[]string
	NamespceList 	[]string
	NsPod			map[string][]string
	AppNameString 	string
	Alertname		string
	Summary			string
	Action			string
	Status			string
	NewNsPod		map[string]map[string]string		// 去重

}

var (
	TOKEN  = os.Getenv("app.env.TOKEN")
	MOBILE = os.Getenv("app.env.MOBILE")
	size AlertSizeInfo
)


//var TOKEN = "https://oapi.dingtalk.com/robot/send?access_token=xxxxxxxx"
//var MOBILE = ""

func ProcessingData(notification Notification) (AlertSizeInfo) {

	var size AlertSizeInfo
	alertname, err := json.Marshal(notification.CommonLabels["alertname"])
	if err != nil {
		log.Println("notification.CommonLabels alertname Marshal failed,", err)
		return size
	} else  {
		size.Alertname = strings.Trim(string(alertname), "\"")
	}

	summary, err := json.Marshal(notification.CommonAnnotations["summary"])
	if err != nil {
		log.Println("notification.CommonAnnotations summary Marshal failed,", err)
		return size
	} else {
		size.Summary = strings.Trim(string(summary), "\"")
	}

	action, err := json.Marshal(notification.CommonLabels["action"])
	if err != nil {
		log.Println("notification.CommonAnnotations action Marshal failed,", err)
		return size
	} else {
		size.Action = strings.Trim(string(action), "\"")
	}

	size.Status = notification.Status

	for i := 0; i < len(notification.Alerts); i++ {
		size.PodList = append(size.PodList, notification.Alerts[i].Annotations["pod"])
		size.AppNameString += " " + notification.Alerts[i].Annotations["pod"]
		size.NamespceList = append(size.NamespceList, notification.Alerts[i].Annotations["namespace"])


		if size.NsPod == nil {
			size.NsPod = make(map[string][]string)
		}

		size.NsPod[notification.Alerts[i].Annotations["namespace"]] = append(size.NsPod[notification.Alerts[i].Annotations["namespace"]], notification.Alerts[i].Annotations["pod"])
		//size.NsPod[notification.Alerts[i].Annotations["namespace"]] = map[string]string{notification.Alerts[i].Annotations["pod"]: "true"}

	}

	for key, _ := range size.NsPod {
		subMap := make(map[string]string)
		for i := 0; i < len(size.NsPod[key]); i++ {
			subMap[size.NsPod[key][i]] = "true"
		}
		if size.NewNsPod == nil {
			size.NewNsPod = make(map[string]map[string]string)
		}
		size.NewNsPod[key] = subMap
	}

	return size
}

// 告警接收
func SendMessage(notification Notification, defaultRobot string, size AlertSizeInfo) {
	var msgres = make(map[string]string)
	msgres["mentioned_mobile_list"] = MOBILE

	// 告警消息
	var buffer bytes.Buffer
	log.Println("开始告警内容后台输出:")
	fmt.Printf("size.Alertname: %s\nsize.Summary: %s\nsize.Action: %s\nsize.Status: %s\n",size.Alertname, size.Summary, size.Action, size.Status)
	fmt.Println("size.NewNsPod:", size.NewNsPod)
	buffer.WriteString(fmt.Sprintf("告警名称: %s\n", size.Alertname))
	buffer.WriteString(fmt.Sprintf("摘要信息: %v\n", size.Summary))
	buffer.WriteString(fmt.Sprintf("触发动作: %v\n", size.Action))
	buffer.WriteString(fmt.Sprintf("当前状态: %v\n", size.Status))
	buffer.WriteString(fmt.Sprintf("以下是相关服务:\n"))
	for key, _ := range size.NewNsPod {
		buffer.WriteString(fmt.Sprintf("命名空间: %s\n", key))
		for podkey, _ := range size.NewNsPod[key] {
			buffer.WriteString(fmt.Sprintf("%-3s异常POD --> : %s\n","*", podkey))
		}

	}

	//buffer.WriteString(fmt.Sprintf("mentioned_mobile_list: %v\n",msgres["mentioned_mobile_list"]))


	// 恢复消息
	if size.Status == "resolved" {
		log.Println("恢复告警内容后台输出:")
		fmt.Printf("size.Alertname: %s\nsize.AppNameString: %s\nsize.Status: %s\n",size.Alertname, size.AppNameString, size.Status)
	}
	var buffer2 bytes.Buffer
	buffer2.WriteString(fmt.Sprintf("尝试恢复服务...\n"))
	buffer2.WriteString(fmt.Sprintf("触发的告警: %s\n", size.Alertname))
	buffer2.WriteString(fmt.Sprintf("当前状态: %v\n",size.Status))
	buffer2.WriteString(fmt.Sprintf("以下是重启的相关服务:\n"))
	for key, _ := range size.NewNsPod {
		buffer2.WriteString(fmt.Sprintf("命名空间: %s\n", key))
		for podkey, _ := range size.NewNsPod[key] {
			buffer2.WriteString(fmt.Sprintf("%-3s重启POD --> : %s\n", "*", podkey))
		}
	}

	var m Message
	m.MsgType = "text"
	m.Text.Mentioned_mobile_list = msgres["mentioned_mobile_list"]

	if size.Status == "resolved" {
		m.Text.Content = buffer2.String()
	}else if size.Status == "firing" {
		m.Text.Content = buffer.String()
	}

	jsons, err := json.Marshal(m)
	if err != nil {
		log.Println("SendMessage Marshal failed,", err)
		return
	}

	resp := string(jsons)
	client := &http.Client{}

	req, err := http.NewRequest("POST", defaultRobot, strings.NewReader(resp))
	if err != nil {
		log.Println("SendMessage http NewRequest failed,", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	r, err := client.Do(req)
	if err != nil {
		log.Println("SendMessage client Do failed", err)
		return
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("SendMessage ReadAll Body failed", err)
		return
	}

	log.Println("SendMessage success,body:", string(body))
}

func ActionDeltePod(size AlertSizeInfo) {
	if strings.ToLower(size.Action) == "deletepod" {
		for ns, _ := range size.NewNsPod {
			for pod, _ :=  range size.NewNsPod[ns] {
				pkg.DeletePod(ns, pod)
			}
		}
	}
}

func Alter(c *gin.Context)  {
	var notification Notification

	err := c.BindJSON(&notification)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//fmt.Println("开始输出notification内容")
	//fmt.Printf("notification类型: %T", notification)
	//fmt.Println("notification.Version: ",notification.Version)
	//fmt.Println("notification.GroupKey: ",notification.GroupKey)
	//fmt.Println("notification.Status: ",notification.Status)
	//fmt.Println("notification.Receiver: ",notification.Receiver)
	//fmt.Println("notification.GroupLabels: ",notification.GroupLabels)
	//fmt.Println("Notification.CommonLabels: ", notification.CommonLabels)
	//fmt.Println("Notification.CommonAnnotations: ", notification.CommonAnnotations)
	//fmt.Println("Notification.ExternalURL: ", notification.ExternalURL)
	//fmt.Println("notification.Alerts长度", len(notification.Alerts))
	//fmt.Println("开始输出notification.Alerts------")
	//for i := 0; i < len(notification.Alerts); i++ {
	//	fmt.Printf("notification.Alerts[%d].Labels: %s\n",i , notification.Alerts[i].Labels)
	//	fmt.Printf("notification.Alerts[%d].Annotations: %s\n",i , notification.Alerts[i].Annotations)
	//	fmt.Println("--------分割线--------")
	//}
	//fmt.Println("notification内容输出完成")

	alertinfo := ProcessingData(notification)

	// 发送告警
	SendMessage(notification, TOKEN, alertinfo)

	// TriggerAction
	ActionDeltePod(alertinfo)

}

func DeleteQuotes(astr string) []string {
	a := "\"ssdfsdf\""
	newstr := strings.Split(a, "")
	newstr = newstr[1 : len(newstr)-1]
	return  newstr
}

func main()  {
	fmt.Printf("TOKEN -> %s\n MOBILE -> %s",TOKEN, MOBILE)
	t := gin.Default()
	t.POST("/Alter",Alter)
	t.Run(":8090")
}
