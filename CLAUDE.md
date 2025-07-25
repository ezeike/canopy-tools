# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based blockchain testing environment for the **Canopy Network** with Ethereum oracle functionality. It simulates cross-chain operations between Canopy blockchain (CNPY token) and Ethereum (USDC token) through an automated oracle system and order book.

## Development Commands

All development commands use the Task runner (`task`). Key commands:

### Environment Setup
- `task anvil` - Start Ethereum testnet with USDC contract deployment
- `task canopy` - Build and start Canopy blockchain node  
- `task node-1/2/3` - Start individual Canopy nodes with hot-reload

### Testing & Automation
- `task e2e` - Run automated end-to-end oracle testing (builds and runs eth-oracle/e2e/)
- `task monitor` - Launch comprehensive monitoring dashboard (balances + orders + storage)
- `task balances` - Monitor ETH, USDC, and CNPY account balances
- `task orders` - Monitor Canopy order book state
- `task store` - Monitor oracle disk storage contents

### Configuration Management  
- `task keygen` - Generate BLS keys for validators
- `task chain-gen` - Generate chain config from eth-oracle profile
- `task chain-gen-default` - Generate default 3-node chain config
- `task chain-clear-data` - Clean all blockchain data directories

### Go Commands
- `go run ./cmd/chain-gen/ <profile>` - Generate chain configurations from templates
- `go run ./cmd/keygen/` - Generate validator keys
- `go build . && go run .` - Build and run E2E tests (from eth-oracle/ directory)

## Architecture Overview

### Core Components
1. **Multi-node Canopy Blockchain** - 3-node network with BLS consensus and committee-based validation
2. **Ethereum Integration** - Local Anvil testnet with USDC ERC20 contract
3. **Oracle System** - Cross-chain order monitoring and transaction processing (node-2 is oracle-enabled)
4. **Order Book System** - CNPY ↔ USDC trading with lock/unlock mechanisms

### Directory Structure
- **cmd/** - CLI utilities (chain-gen, keygen)
- **cmd/keygen** - Canopy BLS key and keystore generator.
- **cmd/chain-gen** - Chain generator - reads YAML templates to generate a node's data-dir
- **eth-oracle/** - Main testing environment and E2E automation
- **chain-profiles/** - YAML network configurations (default.yaml, eth-oracle.yaml)
- **data-dir/** - Runtime blockchain data for each node
- **templates/** - JSON configuration templates for nodes and genesis
- **keys/** - BLS validator keys and Ethereum keystores

### Key Files
- **eth-oracle/e2e/eth_oracle_e2e.go** - Main E2E testing orchestration (1,038 lines)
- **eth-oracle/contracts/USDC.sol** - ERC20 USDC contract for testing
- **taskfile.yml** - Complete task automation and development workflows
- **socket.sh** - Network cleanup script

## Chain Profiles

Two main configurations:
- **default** - 3-node basic setup for general testing
- **eth-oracle** - 2-node setup with oracle functionality (node-2 has oracle: true)

## Environment Variables

The system uses environment files in **eth-oracle/env/**:
- **usdc_contract.env** - USDC contract address and configuration
- **testing.env** - E2E testing parameters

## Testing Workflow

1. Start Ethereum environment: `task anvil`
2. Start Canopy nodes: `task node-1 node-2` (oracle setup) or `task canopy` (single node)
3. Run E2E tests: `task e2e` 
4. Monitor system: `task monitor`

The E2E tests simulate complete cross-chain order lifecycle: order creation → locking on Ethereum → USDC transfer → order completion.

## Hot-Reload Development

Uses Air for hot-reload development with configuration files in **air/** directory. Nodes automatically rebuild on code changes during development.
