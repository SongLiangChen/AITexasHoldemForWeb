//@author: song
//@contact: 462039091@qq.com

package player

import ("hand"
		"card"
		"fmt"
)

type Player struct{
	h *hand.Hand
	chip int
	hole card.Cards
}

func GetPlayer() *Player{
	p := new(Player)
	p.h = hand.GetHand()
	p.chip = 5000
	p.hole = make(card.Cards, 2, 2)
	return p
}

func (p *Player)Init(){
	p.h.Init()
}

func (p *Player)Blind(bet int){
	p.chip -= bet
}

func (p *Player)Call(bet int){
	p.chip -= bet
}

func (p *Player)Take(bet int){
	p.chip += bet
}

func (p *Player)SetHole(c1 *card.Card, c2 *card.Card){
	p.hole[0] = c1
	p.hole[1] = c2
	p.h.SetCard(c1)
	p.h.SetCard(c2)
}

func (p *Player)SetFlop(c1 *card.Card, c2 *card.Card, c3 *card.Card){
	p.h.SetCard(c1)
	p.h.SetCard(c2)
	p.h.SetCard(c3)
}

func (p *Player)SetTurn(c1 *card.Card){
	p.h.SetCard(c1)
}

func (p *Player)SetRiver(c1 *card.Card){
	p.h.SetCard(c1)
}

func (p *Player)DealOver(){
	p.h.AnalyseHand()
}

func (p *Player)ShowChip(){
	fmt.Printf("你的本金还剩：%d\n", p.chip)
}

func (p *Player) GetHoleCards() []int{
	res := make([]int, 2)
	res[0] = p.hole[0].Suit*card.CARDRANK+p.hole[0].Value
	res[1] = p.hole[1].Suit*card.CARDRANK+p.hole[1].Value
	return res
}

func (p *Player)GetChip() int{
	return p.chip
}

func (p *Player)GetLevel() int{
	return p.h.Level
}

func (p *Player)GetFinalValue()int{
	return p.h.FinalValue
}