package main

import ("desk"
		"websocket"
		"fmt"
		"net/http"
		"text/template"
)

func StartGame(ws *websocket.Conn){
	fmt.Println("Echo")
	dd := new(desk.Desk)
	dd.Init(ws)
	dd.PlayGame()
}

func Hello(w http.ResponseWriter, r *http.Request){
	t, _ := template.ParseFiles("3.html")

	t.Execute(w, nil)
}

func GetXml(w http.ResponseWriter, r *http.Request){
	fmt.Println("getXml ",string(desk.GetXmlData()))
	w.Write(desk.GetXmlData())
}

func main(){
	http.Handle("/con", websocket.Handler(StartGame))
	http.HandleFunc("/", Hello)
	http.HandleFunc("/xml", GetXml)

	if err := http.ListenAndServe(":8080", nil); err != nil{
		fmt.Println("LS Fail")
	}
}