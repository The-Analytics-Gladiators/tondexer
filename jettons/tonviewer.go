package jettons

import (
	"TonArb/core"
	"TonArb/models"
	"context"
	"errors"
	"github.com/sethvargo/go-retry"
	"golang.org/x/net/html"
	"log"
	"strconv"
	"strings"
	"time"
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

func parseTokenInfoFromWalletPage(root *html.Node) (*models.TonviewerTokenInfo, error) {
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
	return &models.TonviewerTokenInfo{
		TokenSymbol: tokenSymbol,
		TokenName:   tokenName,
		TokenToUsd:  rate,
	}, nil
}

func TokenInfoFromJettonWalletPage(wallet string) (*models.TonviewerTokenInfo, error) {
	body, e := core.GetRetry(context.Background(), "https://tonviewer.com/"+wallet, 4)
	if e != nil {
		return nil, e
	}

	reader := strings.NewReader(string(body))
	root, e := html.Parse(reader)
	if e != nil {
		return nil, e
	}

	info, e := parseTokenInfoFromWalletPage(root)
	if e != nil {
		return nil, e
	}
	log.Printf("Loaded wallet info: %v, %v, %v \n", info.TokenSymbol, info.TokenName, info.TokenToUsd)
	return info, nil
}

func parseTokenInfoFromMasterPage(root *html.Node) (info *models.TonviewerTokenInfo, er error) {
	defer func() {
		if r := recover(); r != nil {
			info = nil
			er = errors.New("unable to parse html")
		}
	}()

	contractTypeNode := findElementByText(root, "Contract type")
	if contractTypeNode == nil {
		return nil, errors.New("contract type not found")
	}

	parent := contractTypeNode.Parent.NextSibling.FirstChild.FirstChild.NextSibling.FirstChild
	name := parent.
		FirstChild.FirstChild.FirstChild.Data

	symbolWithAmount := parent.
		NextSibling.FirstChild.NextSibling.FirstChild.FirstChild.Data

	ar := strings.SplitN(symbolWithAmount, " ", 2)
	var symbol string
	if len(ar) == 2 {
		symbol = ar[1]
	}

	//rateString := parent.Parent.Parent.NextSibling.FirstChild.FirstChild.FirstChild.Data
	//usdAmount, e := strconv.ParseFloat(strings.Replace(strings.Replace(rateString, ",", "", -1), "$", "", -1), 64)
	//if e != nil {
	//	usdAmount = 0
	//}
	usdAmount := rateFromParent(parent)

	return &models.TonviewerTokenInfo{
		TokenSymbol: symbol,
		TokenName:   name,
		TokenToUsd:  usdAmount,
	}, nil
}

func rateFromParent(parent *html.Node) (f float64) {
	defer func() {
		if r := recover(); r != nil {
			f = 0
		}
	}()

	rateString := parent.Parent.Parent.NextSibling.FirstChild.FirstChild.FirstChild.Data
	usdAmount, e := strconv.ParseFloat(strings.Replace(strings.Replace(rateString, ",", "", -1), "$", "", -1), 64)
	if e != nil {
		usdAmount = 0
	}

	return usdAmount
}

func JettonInfoFromMasterPageRetries(masterAddr string, retries uint64) (*models.TonviewerTokenInfo, error) {
	backoff := retry.WithMaxRetries(retries, retry.NewFibonacci(2*time.Second))
	return retry.DoValue(context.Background(), backoff, func(ctx context.Context) (*models.TonviewerTokenInfo, error) {
		jettonInfo, err := JettonInfoFromMasterPage(masterAddr)
		return jettonInfo, retry.RetryableError(err)
	})
}

func JettonInfoFromMasterPage(master string) (*models.TonviewerTokenInfo, error) {
	body, e := core.GetRetry(context.Background(), "https://tonviewer.com/"+master, 4)
	if e != nil {
		return nil, e
	}

	reader := strings.NewReader(string(body))
	root, e := html.Parse(reader)
	if e != nil {
		return nil, e
	}
	return parseTokenInfoFromMasterPage(root)
}
