//@author: song
//@contact: 462039091@qq.com

package desk

import ("fmt"
		"ai"
		"dealmachine"
		//"card"
		"player"
		//"hand"
		"encoding/xml"
		"websocket"
)

var CARDTYPE = []string{"", "高牌", "一对", "两对", "三条", "顺子", "同花", "葫芦", "四条", "同花顺", "皇家同花顺"}

const (
	STATE_INIT = 0
	STATE_HOLE = 1
	STATE_FLOP = 2
	STATE_TURN = 3
	STATE_RIVER = 4
)

type HTMLData struct{
	XMLName xml.Name     `xml:"htmldata"`
	AiChip int           `xml:"aichip"`
	GambPool int         `xml:"gambpool"`
	CommityCards string  `xml:"commitycards"`
	PlayerChip int       `xml:"playerchip"`
	PlayerHold string    `xml:"playerhole"`
	Prompt string        `xml:"prompt"`
}

type BetInfo struct{
	XMLName  xml.Name `xml:"root"`
	Bet int `xml:"bet"`
	Call int `xml:"call"`
	Fold int `xml:"fold"`
}

type Desk struct{
	gambpool int
	current int

	dm *dealmachine.DealMachine
	aai *ai.AI
	p *player.Player
	dd *HTMLData

	playerAllIn bool
	aiAllin bool

	state int
	ws *websocket.Conn

	offline bool
}

var xmlData string
func GetXmlData()[]byte{
	return []byte(xmlData)
}

func (d *Desk)IsOffline() bool{
	return d.offline
}

func (d *Desk)receive()BetInfo{
	var reply string
	if err := websocket.Message.Receive(d.ws, &reply); err != nil{
		fmt.Println(err)
		d.offline = true
		return BetInfo{}
	}
	if(reply == "13"){
		return BetInfo{}
	}
	b := new(BetInfo)
	fmt.Println("recive ", reply)
	err := xml.Unmarshal([]byte(reply), b)
	if err != nil{
		fmt.Println(err)
	}
	return *b
}

func (d *Desk)send() bool{

	var msg []byte
	var err error
	if msg, err = xml.Marshal(*d.dd); err != nil{
		return false
	}
	fmt.Println("xml ",string(msg))
	xmlData = string(msg)
	if err = websocket.Message.Send(d.ws, xmlData); err != nil{
		d.offline = true
		return false
	}
	return true
}

func (d *Desk)Init(wws *websocket.Conn){
	d.dm = dealmachine.GetDealMachine()
	d.dm.Init()

	d.aai = ai.GetAI()
	d.aai.Init()
	d.p = player.GetPlayer()
	d.p.Init()

	d.dd = new(HTMLData)
	d.current = 0
	d.state = STATE_HOLE
	d.ws = wws
	d.offline = false
}

func (d *Desk)handHole(){
	d.dm.Shuffle()
	d.aai.Init()
	d.p.Init()

	if d.p.GetChip()<=0 || d.aai.GetChip()<=0{
		d.offline = true
		return
	}
	d.gambpool = 0
	d.aiAllin = false
	d.playerAllIn = false

	d.p.Blind(100)
	d.aai.Blind(100)
	d.gambpool += 200

	d.p.SetHole(d.dm.Deal(), d.dm.Deal())
	d.aai.SetHole(d.dm.Deal(), d.dm.Deal())

	d.state = STATE_FLOP

	if d.p.GetChip()==0 || d.aai.GetChip()==0{
		d.allin()
		d.state = STATE_HOLE
		return
	} 
	bet := d.chipIn()
	if bet == 0 || d.offline == true{
		d.state = STATE_HOLE
		return
	}

	d.gambpool += 2*bet
	d.p.Call(bet)
	d.aai.Call(bet)

	if d.playerAllIn==true || d.aiAllin==true{
		d.allin()
		d.state = STATE_HOLE
	}
}

func (d *Desk)handFlop(){
	c1 := d.dm.Deal()
	c2 := d.dm.Deal()
	c3 := d.dm.Deal()

	d.p.SetFlop(c1, c2, c3)
	d.aai.SetFlop(c1, c2, c3)

	d.state = STATE_TURN

	bet := d.chipIn()
	if bet == 0 || d.offline == true{
		d.state = STATE_HOLE
		return
	}

	d.gambpool += 2*bet
	d.p.Call(bet)
	d.aai.Call(bet)

	if d.playerAllIn==true || d.aiAllin==true{
		d.allin()
		d.state = STATE_HOLE
	}
}

func (d *Desk)handTurn(){
	c := d.dm.Deal()

	d.p.SetTurn(c)
	d.aai.SetTurn(c)

	d.state = STATE_RIVER

	bet := d.chipIn()
	if bet == 0 || d.offline == true{
		d.state = STATE_HOLE
		return
	}

	d.gambpool += 2*bet
	d.p.Call(bet)
	d.aai.Call(bet)

	if d.playerAllIn==true || d.aiAllin==true{
		d.allin()
		d.state = STATE_HOLE
	}
}

func (d *Desk)handRiver(){
	c := d.dm.Deal()

	d.p.SetRiver(c)
	d.aai.SetRiver(c)

	d.state = STATE_HOLE

	bet := d.chipIn()
	if bet == 0 || d.offline == true{
		return
	}

	d.gambpool += 2*bet
	d.p.Call(bet)
	d.aai.Call(bet)

	d.showdown()
}

func (d *Desk)PlayGame(){

	for {
		if d.offline == true{
			d.packageData()
			d.dd.Prompt = "游戏结束 "
			if d.p.GetChip() > 0{
				d.dd.Prompt += "你真牛逼"
			}else{
				d.dd.Prompt += "你弱爆了"
			}
			d.send()
			return
		}

		switch d.state{
			case STATE_HOLE:
				d.handHole()
				break
			case STATE_FLOP:
				d.handFlop()
				break
			case STATE_TURN:
				d.handTurn()
				break
			case STATE_RIVER:
				d.handRiver()
				break
		}
	}
}

func (d *Desk)chipIn() int{

	var fcr int
	var bet int
	var betinfo BetInfo
	if d.current == 0{
		d.packageData()
		d.dd.Prompt = "玩家请下注"
		d.send()
		if d.offline == true{
			return 0
		}
		for{
			betinfo = d.receive()
			if d.offline == true{
				return 0
			}
			if betinfo.Fold == 1 || betinfo.Bet > 0{
				break
			}
			d.send()
		}
		if betinfo.Fold == 1{
			playerfold(d)
			return 0
		}
		
		bet = betinfo.Bet

		if bet >= d.p.GetChip(){
			bet = d.p.GetChip()
			d.playerAllIn = true
		}
		fcr = d.aai.FCR(bet, d.gambpool, d.dm)
		if fcr == 0{
			aifold(d)
			return 0
		}
		if fcr >= d.aai.GetChip(){
			fcr = d.aai.GetChip()
			d.aiAllin = true
		}
		if d.playerAllIn == false && fcr > bet{
			d.current = 1
			d.packageData()
			d.dd.Prompt = fmt.Sprintf("电脑加注为：%d,是否跟注？", fcr)
			d.send()
			if d.offline == true{
				return 0
			}

			for {
				betinfo = d.receive()
				if d.offline == true{
					return 0
				}
				if betinfo.Call == 1 || betinfo.Fold == 1{
					break
				}
				d.send()
			}
			if betinfo.Fold == 1{
				playerfold(d)
				return 0
			}
			if fcr >= d.p.GetChip(){
				bet = d.p.GetChip()
				d.playerAllIn = true
			}else{
				bet = fcr
			}
		}
	}else{
		fcr = d.aai.FCR(0, d.gambpool, d.dm)
		if fcr == 0{
			aifold(d)
			return 0
		}
		if fcr >= d.aai.GetChip(){
			fcr = d.aai.GetChip()
			d.aiAllin = true
		}
		d.packageData()
		d.dd.Prompt = fmt.Sprintf("电脑加注为：%d,请加注或者跟注,加注不能小于%d", fcr, fcr)
		d.send()
		if d.offline == true{
			return 0
		}

		for {
			betinfo = d.receive()
			if d.offline == true{
				return 0
			}
			if betinfo.Call == 1 || betinfo.Fold == 1 || betinfo.Bet >= fcr{
				break
			}
			d.send()
		}
		if betinfo.Fold == 1{
			playerfold(d)
			return 0
		}
		if betinfo.Call == 1{
			bet = fcr
		}else{
			bet = betinfo.Bet
		}
		
		if bet >= d.p.GetChip(){
			bet = d.p.GetChip()
			d.playerAllIn = true
		}

		if fcr < bet && d.aiAllin == false{
			tmp := d.aai.FCR(bet, d.gambpool+bet, d.dm)
			if tmp==0 {
				aifold(d)
				return 0
			}
			if bet>=d.aai.GetChip(){
				fcr = d.aai.GetChip()
				d.aiAllin = true
			}else{
				fcr = bet
			}
		}
	}
	if d.aiAllin == true || d.playerAllIn == true{
		if bet < fcr{
			return bet
		}
		return fcr
	}
	return bet
}

func (d *Desk)allin(){
	if d.state == STATE_FLOP{
		c1 := d.dm.Deal()
		c2 := d.dm.Deal()
		c3 := d.dm.Deal()
		d.p.SetFlop(c1, c2, c3)
		d.aai.SetFlop(c1, c2, c3)
		d.state = STATE_TURN
	}
	if d.state == STATE_TURN{
		c := d.dm.Deal()
		d.p.SetTurn(c)
		d.aai.SetTurn(c)
		d.state = STATE_RIVER
	}
	if d.state == STATE_RIVER{
		c := d.dm.Deal()
		d.p.SetRiver(c)
		d.aai.SetRiver(c)
	}
	d.showdown()
}

func (d *Desk)showdown(){
	d.p.DealOver()
	d.aai.DealOver()

	if d.p.GetLevel() > d.aai.GetLevel(){
		playerwin(d)
	}else if d.p.GetLevel() == d.aai.GetLevel() && d.p.GetFinalValue() > d.aai.GetFinalValue(){
		playerwin(d)
	}else if d.p.GetLevel() == d.aai.GetLevel() && d.p.GetFinalValue() == d.aai.GetFinalValue(){
		playertie(d)
	}else{
		playerfail(d)
	}
	d.state = STATE_HOLE
	fmt.Println("copy")
}

func checkError(e error){
	if e != nil{
		fmt.Println(e)
	}
}

func (d *Desk)packageData(){
	d.dd.AiChip = d.aai.GetChip()
	d.dd.GambPool = d.gambpool
	d.dd.CommityCards = d.aai.GetCommityCards()
	d.dd.PlayerChip = d.p.GetChip()
	d.dd.PlayerHold = d.p.GetHoleCards()
	d.dd.Prompt = ""
}

func playerfold(d *Desk){
	d.current = 1
	d.aai.Take(d.gambpool)
	d.gambpool = 0
	d.packageData()
	d.dd.Prompt = "玩家弃牌，电脑获胜，点击重新发牌继续"
	d.send()
	d.receive()
}

func aifold(d *Desk){
	d.current = 0
	d.p.Take(d.gambpool)
	d.gambpool = 0
	d.packageData()
	d.dd.Prompt = "电脑弃牌，玩家获胜，点击重新发牌继续"
	d.send()
	d.receive()
}

func playerwin(d *Desk){
	d.current = 0
	d.p.Take(d.gambpool)
	d.gambpool = 0
	d.packageData()
	d.dd.Prompt = "玩家牌大，电脑手牌为："
	d.dd.Prompt += d.aai.GetHole()
	d.dd.Prompt += "  牌型为："
	d.dd.Prompt += CARDTYPE[d.aai.GetLevel()]
	d.dd.Prompt += "  点击重新发牌继续"
}

func playerfail(d *Desk){
	d.current = 1
	d.aai.Take(d.gambpool)
	d.gambpool = 0
	d.packageData()
	d.dd.Prompt = "电脑牌大，电脑手牌为："
	d.dd.Prompt += d.aai.GetHole()
	d.dd.Prompt += "  牌型为："
	d.dd.Prompt += CARDTYPE[d.aai.GetLevel()]
	d.dd.Prompt += "  点击重新发牌继续"
	d.send()
	d.receive()
}

func playertie(d *Desk){
	d.p.Take(d.gambpool/2)
	d.aai.Take(d.gambpool/2)
	d.gambpool = 0
	d.packageData()
	d.dd.Prompt = "平局，电脑手牌为："
	d.dd.Prompt += d.aai.GetHole()
	d.dd.Prompt += "  点击重新发牌继续"
	d.send()
	d.receive()
}
