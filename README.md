# CDK Data Availability

### Data Availability Layer for CDK Validium

The cdk-data-availability project is a specialized Data Availability Node (DA Node) that is part of Polygon's CDK (Chain Development Kit) Validium.

## Overview of Validium

For a full overview of the Polygon CDK Validium, please reference the [CDK documentation](https://wiki.polygon.technology/docs/cdk/).

The CDK Validium solution is made up of several components; start with the [CDK Validium Node](https://github.com/0xPolygon/cdk-validium-node). For quick reference, the complete list of components are outlined below:

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

Data Availability Nodes facilitate a separation between transaction execution and data storage. They allow transaction data to reside off-chain while remaining accessible for validation. This significantly improves scalability and reduces costs. Within the framework of Polygon's CDK, Data Availability Committees (DAC) members run DA nodes to ensure the security, accessibility, and reliability of off-chain data.

To learn more about how the data availability layer works in the validium, please see the CDK documentation [here](https://wiki.polygon.technology/docs/cdk/dac-overview/).

### Off-Chain Data

The off-chain data is stored in a distributed manner and managed by a data availability committee, ensuring that it is available for validation. The data availability committee is defined as a core smart contract, available [here](https://github.com/0xPolygon/cdk-validium-contracts/blob/main/contracts/CDKDataCommittee.sol). This is crucial for the Validium model, where data computation happens off-chain but needs to be verifiable on-chain.

### Running

Instructions on how to run this software can be found [here](./docs/running.md)

## License

The cdk-validium-node project is licensed under the [GNU Affero General Public License](LICENSE) free software license.
