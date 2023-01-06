run:
	go run main.go

sol:
	# solcjs --optimize --abi ./contracts/Brussels.sol -o build
	solcjs --optimize --bin ./contracts/Brussels.sol -o build
	abigen --abi=./build/contracts_Brussels_sol_Brussels.abi --bin=./build/contracts_Brussels_sol_Brussels.bin --pkg=api --out=./api/smart_contract.go