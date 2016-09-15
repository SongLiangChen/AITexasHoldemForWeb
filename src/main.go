package main

import ("desk"
		"websocket"
		"fmt"
		"net/http"
		"text/template"
		//"os"
)

func StartGame(ws *websocket.Conn){
	fmt.Println("new connection")
	dd := new(desk.Desk)
	dd.Init(ws)
	dd.PlayGame()
}

func Hello(w http.ResponseWriter, r *http.Request){
	t, _ := template.ParseFiles("3.html")

	t.Execute(w, nil)
}

func GetXml(w http.ResponseWriter, r *http.Request){
	//fmt.Println("getXml ",string(desk.GetXmlData()))
	w.Write(desk.GetXmlData())
}

func GetImg(w http.ResponseWriter, r *http.Request){
	//fmt.Println("getimg")
	r.ParseForm()
	imageId := r.FormValue("id")
	imagePath := "sourse" + "/" + imageId
	w.Header().Set("Content-Type", "image")
	http.ServeFile(w, r, imagePath)
}

func main(){
	http.Handle("/con", websocket.Handler(StartGame))
	http.HandleFunc("/", Hello)
	http.HandleFunc("/xml", GetXml)
	http.HandleFunc("/sourse", GetImg)

	if err := http.ListenAndServe(":8080", nil); err != nil{
		fmt.Println("LS Fail")
	}
	fmt.Println("Start")
}
