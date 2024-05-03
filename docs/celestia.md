
# Setting Up A Light Node with Celestia

The following guide provides a comprehensive introduction to setting up a light node with Celestia, a robust and reliable data availability layer for blockchains.
We'll go through the necessary hardware requirements, explain how to install the main Celestia components (the celestia-node and the celestia-app), detail the steps to initialize and start the light node, and provide instructions on handling keys and wallets. This allows you to be a part of the Celestia network and participate in maintaining data availability. This guide is especially useful for those who are looking to understand the steps involved in running a Celestia light node and the commands required to set it up successfully.

Let's guide you through the process of becoming a part of the Celestia network by running your own light node.


### Minimum Hardware requirements

The following minimum hardware requirements are recommended for running a light node:

* Memory: 500 MB RAM (minimum)
* CPU: Single Core 
* Disk: 50 GB SSD Storage
* Bandwidth: 56 Kbps for Download/56 Kbps for Upload



### Install Celestia-Node and Celestia-App


To start, install both the Celestia-node and Celestia-app packages. This step-by-step guide makes installation a breeze.

#### Install celestia-node

`bash -c "$(curl -sL https://docs.celestia.org/celestia-node.sh)"`

#### Install celestia-app
`bash -c "$(curl -sL https://docs.celestia.org/celestia-app.sh)"`



### Initialize the light node


`celestia light init --p2p.network mocha`


### Keys and wallets

You can create your key for your node by running the following command with the `cel-key` utility in the `celestia-node` directory:

`./cel-key add <key-name> --keyring-backend test  --node.type light --p2p.network mocha`


### Start The Celestia Light Client
Once your key is created, you're ready to start the Celestia light client. Use the following command, specifying the key name you created earlier:

`celestia light start --keyring.accname my_celes_key --core.ip rpc-mocha.pops.one --p2p.network mocha`