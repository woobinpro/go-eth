package main

import (
	"fmt"
	"github.com/ethereum/eth-go"
	"github.com/ethereum/eth-go/ethchain"
	"github.com/ethereum/eth-go/ethminer"
	"github.com/ethereum/eth-go/ethutil"
	"github.com/ethereum/go-ethereum/utils"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strings"
)

const Debug = true

// Register interrupt handlers so we can stop the ethereum
func RegisterInterupts(s *eth.Ethereum) {
	// Buffered chan of one is enough
	c := make(chan os.Signal, 1)
	// Notify about interrupts for now
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Printf("Shutting down (%v) ... \n", sig)
			s.Stop()
		}
	}()
}

func main() {
	Init()

	runtime.GOMAXPROCS(runtime.NumCPU())

	// set logger
	var logger *log.Logger
	flags := log.LstdFlags

	if LogFile != "" {
		logfile, err := os.OpenFile(LogFile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
    if err != nil {
      panic(fmt.Sprintf("error opening log file '%s': %v", LogFile, err))
    }
    defer logfile.Close()
		log.SetOutput(logfile)
		logger = log.New(logfile, "", flags)
  } else {
  	logger = log.New(os.Stdout, "", flags)
  }
	ethutil.ReadConfig(DataDir)
	ethutil.Config.Log.AddLogSystem(logger)

  ethchain.InitFees()
	ethutil.Config.Seed = UseSeed

	// Instantiated a eth stack
	ethereum, err := eth.New(eth.CapDefault, UseUPnP)
	if err != nil {
		log.Println("eth start err:", err)
		return
	}
	ethereum.Port = OutboundPort

	if GenAddr {
		fmt.Println("This action overwrites your old private key. Are you sure? (y/n)")

		var r string
		fmt.Scanln(&r)
		for ; ; fmt.Scanln(&r) {
			if r == "n" || r == "y" {
				break
			} else {
				fmt.Printf("Yes or no?", r)
			}
		}

		if r == "y" {
			utils.CreateKeyPair(true)
		}
		os.Exit(0)
	} else {
		if len(ImportKey) > 0 {
			fmt.Println("This action overwrites your old private key. Are you sure? (y/n)")
			var r string
			fmt.Scanln(&r)
			for ; ; fmt.Scanln(&r) {
				if r == "n" || r == "y" {
					break
				} else {
					fmt.Printf("Yes or no?", r)
				}
			}

			if r == "y" {
				mnemonic := strings.Split(ImportKey, " ")
				if len(mnemonic) == 24 {
					fmt.Println("Got mnemonic key, importing.")
					key := ethutil.MnemonicDecode(mnemonic)
					utils.ImportPrivateKey(key)
				} else if len(mnemonic) == 1 {
					fmt.Println("Got hex key, importing.")
					utils.ImportPrivateKey(ImportKey)
				} else {
					fmt.Println("Did not recognise format, exiting.")
				}
				os.Exit(0)
			}
		} else {
			utils.CreateKeyPair(false)
		}
	}

	if ExportKey {
		key := ethutil.Config.Db.GetKeys()[0]
		fmt.Printf("%x\n", key.PrivateKey)
		os.Exit(0)
	}

	if ShowGenesis {
		fmt.Println(ethereum.BlockChain().Genesis())
		os.Exit(0)
	}

	ethutil.Config.Log.Infoln(fmt.Sprintf("Starting Ethereum v%s", ethutil.Config.Ver))

	// Set the max peers
	ethereum.MaxPeers = MaxPeer

	if StartConsole {
		err := os.Mkdir(ethutil.Config.ExecPath, os.ModePerm)
		// Error is OK if the error is ErrExist
		if err != nil && !os.IsExist(err) {
			log.Panic("Unable to create EXECPATH:", err)
		}

		console := NewConsole(ethereum)
		go console.Start()
	}

	RegisterInterupts(ethereum)
	ethereum.Start()

	if StartMining {
		ethutil.Config.Log.Infoln("Miner started")

		// Fake block mining. It broadcasts a new block every 5 seconds
		go func() {

			if StartMining {
				ethutil.Config.Log.Infoln("Miner started")

				go func() {
					data, _ := ethutil.Config.Db.Get([]byte("KeyRing"))
					keyRing := ethutil.NewValueFromBytes(data)
					addr := keyRing.Get(1).Bytes()

					miner := ethminer.NewDefaultMiner(addr, ethereum)
					miner.Start()

				}()
			}
		}()

	}

	// Wait for shutdown
	ethereum.WaitForShutdown()
}
