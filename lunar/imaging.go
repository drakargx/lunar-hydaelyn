package lunar

import (
	"fmt"
	"sort"

	"github.com/beevik/etree"
)

const (
	templateSVG   = "template.svg"
	placeholderID = "PLACEHOLDER"
)

type SortByRdps []RdpsRankingCharacter

func (r SortByRdps) Len() int {
	return len(r)
}

func (r SortByRdps) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func (r SortByRdps) Less(i, j int) bool {
	return r[i].Rdps > r[j].Rdps
}

func GenerateOutputPng(inkscapeLocation string, structure DeconstructedQueryResponse) {

	//possibly break DeconstructedQueryResponse into a map[string]string that looks like
	//"PLAYER0": structure.RdpsRankings[0].PlayerName
	//"DPS0": structure.RdpsRankings[0].Rdps
	//"PLAYER1": structure.RdpsRankings[1].PlayerName
	//etc..

	players := make([]RdpsRankingCharacter, len(structure.Players))
	j := 0
	for _, v := range structure.Players {
		players[j] = v
		j++
	}
	sort.Sort(SortByRdps(players))

	playerMapping := make(map[string]string)

	for i := 0; i < len(players); i++ {
		playerName := players[i].PlayerName
		playerMapping[fmt.Sprint("PLAYER", i)] = playerName
		playerMapping[fmt.Sprint("JOB", i)] = structure.Jobs[playerName]
		playerMapping[fmt.Sprint("DPS", i)] = fmt.Sprintf("%.2f", players[i].Rdps)
		//playerMapping[fmt.Sprint("DD", i)] = fmt.Sprintf("%d", structure.DamageDowns[playerName])
	}

	doc := etree.NewDocument()
	doc.ReadFromFile(templateSVG)

	root := doc.SelectElement("svg")
	for _, node := range root.SelectElements("g") {
		if idNode := node.SelectAttr("id"); idNode != nil && idNode.Value == placeholderID {
			for _, attrNode := range node.ChildElements() {

				idName := attrNode.SelectAttrValue("id", "noid")

				attrNode.SelectElement("text").SetText(playerMapping[idName])

			}

			break
		}
	}

	doc.WriteToFile("Output.svg")
}
