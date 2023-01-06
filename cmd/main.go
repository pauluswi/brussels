package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	//"02amanag/bc/api" // this would be your generated smart contract bindings
	"brussels/api"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// address of etherum env
	client, err := ethclient.Dial("http://127.0.0.1:7545")
	if err != nil {
		panic(err)
	}

	// create auth and transaction package for deploying smart contract
	//auth := getAccountAuth(client, "fd4eef6dec5575cc78f3f14d4b749094f8b88ad7883caaa8d1d24e9a01e3732d")
	auth := getAccountAuth(client, "ea9816250aa83b4dd8d265f9cde9acbc0b2599a0864e907aa84cbe68cd9d26e4")

	// 0x60C749BcaD846C837a27Ec77B193Fa5c104018E0
	// 85512b9a6c3a48a5dfa8bff67e9a95bd3313bc21b36a6283137c8842a6c380f2
	//auth := getAccountAuth(client, "85512b9a6c3a48a5dfa8bff67e9a95bd3313bc21b36a6283137c8842a6c380f2")

	//deploying smart contract
	address, tx, instance, err := api.DeployApi(auth, client)
	if err != nil {
		panic(err)
	}

	fmt.Println(address.Hex())

	_, _ = instance, tx
	fmt.Println("instance->", instance)
	fmt.Println("tx->", tx.Hash().Hex())

	//creating api object to intract with smart contract function
	conn, err := api.NewApi(common.HexToAddress(address.Hex()), client)
	if err != nil {
		panic(err)
	}

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/balance", func(c echo.Context) error {
		reply, err := conn.Balance(&bind.CallOpts{})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, reply)
	})
	e.GET("/admin", func(c echo.Context) error {
		reply, err := conn.Admin(&bind.CallOpts{})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, reply)
	})
	e.POST("/deposite/:amount", func(c echo.Context) error {
		amount := c.Param("amount")
		amt, _ := strconv.Atoi(amount)

		//gets address of account by which amount to be deposite
		var v map[string]interface{}
		err := json.NewDecoder(c.Request().Body).Decode(&v)
		if err != nil {
			panic(err)
		}

		//creating auth object for above account
		auth := getAccountAuth(client, v["accountPrivateKey"].(string))

		reply, err := conn.Deposite(auth, big.NewInt(int64(amt)))
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, reply)
	})
	e.POST("/withdrawl/:amount", func(c echo.Context) error {
		amount := c.Param("amount")
		amt, _ := strconv.Atoi(amount)

		var v map[string]interface{}
		err := json.NewDecoder(c.Request().Body).Decode(&v)
		if err != nil {
			panic(err)
		}

		auth := getAccountAuth(client, v["accountPrivateKey"].(string))
		// auth.Nonce.Add(auth.Nonce, big.NewInt(int64(1))) //it is use to create next nounce of account if it has to make another transaction

		reply, err := conn.Withdrawl(auth, big.NewInt(int64(amt)))
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, reply)
	})

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

//function to create auth for any account from its private key
func getAccountAuth(client *ethclient.Client, privateKeyAddress string) *bind.TransactOpts {

	privateKey, err := crypto.HexToECDSA(privateKeyAddress)
	if err != nil {
		panic(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("invalid key")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		panic(err)
	}
	fmt.Println("nounce=", nonce)
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		panic(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		panic(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(3000000) // in units
	auth.GasPrice = big.NewInt(1000000)

	return auth
}
