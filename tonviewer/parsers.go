package tonviewer

import (
	"TonArb/models"
	"errors"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func findElementByText(n *html.Node, searchText string) *html.Node {
	if n.Type == html.TextNode && strings.TrimSpace(n.Data) == searchText {
		return n.Parent // Return the parent node containing the text
	}
	// Recursively check child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		found := findElementByText(c, searchText)
		if found != nil {
			return found
		}
	}
	return nil
}

func parseTokenInfo(root *html.Node) (*models.TokenInfo, error) {
	amountNode := findElementByText(root, "Amount")
	if amountNode == nil {
		return nil, errors.New("amount not found")
	}
	amountsParent := amountNode.NextSibling

	tokenAmountParent := amountsParent.FirstChild
	tokenAmountString := tokenAmountParent.FirstChild.Data

	tokenAmount, e := strconv.ParseFloat(strings.Replace(tokenAmountString, ",", "", -1), 64)
	if e != nil {
		return nil, e
	}

	tokenSymbol := tokenAmountParent.FirstChild.NextSibling.NextSibling.NextSibling.NextSibling.Data

	usdAmountString := tokenAmountParent.NextSibling.FirstChild.NextSibling.NextSibling.Data

	usdAmount, e := strconv.ParseFloat(strings.Replace(strings.Replace(usdAmountString, ",", "", -1), "$", "", -1), 64)
	if e != nil {
		return nil, e
	}

	contractTypeNode := findElementByText(root, "Contract type")
	tokenInfoNode := contractTypeNode.Parent.NextSibling

	tokenName := tokenInfoNode.FirstChild.NextSibling.FirstChild.FirstChild.FirstChild.FirstChild.Data

	rate := usdAmount / tokenAmount
	return &models.TokenInfo{
		TokenSymbol: tokenSymbol,
		TokenName:   tokenName,
		TokenToUsd:  rate,
	}, nil
}

func FetchTokenInfo(token string) (*models.TokenInfo, error) {
	resp, e := http.Get("https://tonviewer.com/" + token)
	if e != nil {
		return nil, e
	}
	defer resp.Body.Close()

	root, e := html.Parse(resp.Body)
	if e != nil {
		return nil, e
	}

	info, e := parseTokenInfo(root)
	if e != nil {
		return nil, e
	}
	log.Printf("Loaded token info: %v, %v, %v \n", info.TokenSymbol, info.TokenName, info.TokenToUsd)
	return info, nil
}
