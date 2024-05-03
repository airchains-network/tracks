#!/bin/bash
echo "Setting Up A Light Node with Celestia"


echo -e "\n*** Installing celestia-node ***"
bash -c "$(curl -sL https://docs.celestia.org/celestia-node.sh)"

echo -e "\n*** Installing celestia-app ***"
bash -c "$(curl -sL https://docs.celestia.org/celestia-app.sh)"

echo -e "\n*** Initializing the light node ***"
celestia light init --p2p.network mocha

echo -e "\n*** Creating keys and wallets ***"
echo -e "Note: Replace 'key-name' with your desired key name in the command below."
./cel-key add <key-name> --keyring-backend test  --node.type light --p2p.network mocha

echo -e "\n*** Starting The Celestia Light Client ***"
echo -e "Note: Replace 'my_celes_key' with the key name you created above in the command below."
celestia light start --keyring.accname my_celes_key --core.ip rpc-mocha.pops.one --p2p.network mocha

echo -e "\n*** Completed Setting up the Celestia Light Node ***"