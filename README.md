<div align="center">
<h1>CDK Data Availability</h1>
<h3>Data Availability Layer for CDK Validium</h3>

</div>

<p align="left">
  The cdk-data-availability project is a specialized Data Availability Node (DA Node) that is part of Polygon's CDK (Chain Development Kit) Validium.
</p>

<!-- TOC -->

- [Overview of Validium](#overview-of-validium)
- [Introduction](#introduction)
- [Key Components](#key-components)
  * [Off-Chain Data](#off-chain-data)
  * [Data Availability Committee](#data-availability-committee)
  * [Sequencer](#sequencer)
  * [Synchronizer](#synchronizer)
- [Prerequisites](#prerequisites)
- [Deployment](#deployment)
- [License](#license)

## Overview of Validium

For a full overview of the CDK-Validium please reference the [CDK documentation](https://wiki.polygon.technology/docs/cdk/).

The CDK-Validium solution is made up of several components, start with the [CDK Validium Node](https://github.com/0xPolygon/cdk-validium-node). For quick reference, the complete list of components are outlined below:

| Component                                                                     | Description                                                          |
| ----------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| [CDK Validium Node](https://github.com/0xPolygon/cdk-validium-node)           | Node implementation for the CDK networks in Validium mode            |
| [CDK Validium Contracts](https://github.com/0xPolygon/cdk-validium-contracts) | Smart contract implementation for the CDK networks in Validium mode |
| [CDK Data Availability](https://github.com/0xPolygon/cdk-data-availability)   | Data availability implementation for the CDK networks          |
| [Prover / Executor](https://github.com/0xPolygonHermez/zkevm-prover)          | zkEVM engine and prover implementation                               |
| [Bridge Service](https://github.com/0xPolygonHermez/zkevm-bridge-service)     | Bridge service implementation for CDK networks                       |
| [Bridge UI](https://github.com/0xPolygonHermez/zkevm-bridge-ui)               | UI for the CDK networks bridge                                       |

---

## Introduction

As blockchain networks grow, the volume of data that needs to be stored and validated increases, posing challenges in scalability and efficiency. Storing all data on-chain can lead to bloated blockchains, slow transactions, and high fees.

Data Availability Nodes facilitate a separation between transaction execution and data storage. They allow transaction data to reside off-chain while remaining accessible for validation. This significantly improves scalability and reduces costs. Within the framework of Polygon's CDK, the DA Node is managed by a Data Availability Committee (DAC) to ensure the security, accessibility, and reliability of off-chain data.

To learn more about how the data availability layer works in the validium, please see the CDK documentation [here](https://wiki.polygon.technology/docs/cdk/dac-overview/).

## Key Components

### Off-Chain Data

The off-chain data is stored in a distributed manner and managed by a data availability committee, ensuring that it is available for validation. The data availability committee is defined as a core smart contract, available [here]. This is crucial for the Validium model, where data computation happens off-chain but needs to be verifiable on-chain.

### Data Availability Committee

This is a set of nodes responsible for ensuring that the off-chain data is available and can be validated by any network participant. They work in conjunction with the sequencer to maintain data integrity.

### Sequencer

The sequencer is responsible for ordering the data that the sequencer will send to L1. It also handles the accumulated input hash (`accInputHash`), which is crucial for data integrity.

### Synchronizer

The synchronizer package has several key components:

- **BatchSynchronizer**: Manages data synchronization by watching for batch events on the blockchain.
- **ReorgDetector**: Monitors for block reorganizations and alerts other components to adapt to the new chain state.
- **Watcher**: A utility component for subscribing to blockchain events.

## Prerequisites

- **Go (Check by running `go version`)**
- **curl (Check by running `curl --version`)**
- **Docker (Check by running `docker --version`)**

## Deployment

1. Clone the repository:

    ```bash
    git clone https://github.com/0xPolygon/cdk-data-availability.git
    ```

2. Navigate to the project directory:

    ```bash
    cd cdk-data-availability
    ```

3. Check for Required Dependencies:

    ```bash
    make check-go
    make check-curl
    make check-docker
    ```

4. Build the Binary:

    ```bash
    make build
    ``````

5. Build Docker Image:

    ```bash
    make build-docker
    ```

## License

The cdk-validium-node project is licensed under the [GNU Affero General Public License](LICENSE) free software license.
