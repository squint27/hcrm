#!/bin/bash
# Helper script for running integration tests against Hetzner Cloud API
# This script provides a user-friendly interface for running tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_help() {
    cat << EOF
${BLUE}Hetzner Cloud Integration Test Helper${NC}

Usage: ./scripts/integration-tests.sh [COMMAND] [OPTIONS]

${BLUE}Commands:${NC}
  run           Run all integration tests
  run-single    Run a specific integration test
  check-token   Check if HCLOUD_TOKEN is set
  help          Show this help message

${BLUE}Options:${NC}
  --token TOKEN     Set Hetzner Cloud API token
  --timeout SECS    Test timeout in seconds (default: 300)
  --parallel N      Number of parallel tests (default: 1)
  --verbose         Enable verbose output
  --race            Enable race condition detection

${BLUE}Examples:${NC}
  # Run all tests with your token
  export HCLOUD_TOKEN="your-token"
  ./scripts/integration-tests.sh run

  # Run single test
  ./scripts/integration-tests.sh run-single TestHCloudIntegration_CreateAndGetNetwork

  # Check if token is set
  ./scripts/integration-tests.sh check-token

  # Run with custom options
  ./scripts/integration-tests.sh run --timeout 600 --verbose

${BLUE}Environment Variables:${NC}
  HCLOUD_TOKEN      Hetzner Cloud API token (required for tests)
  GO_TEST_VERBOSE   Enable verbose output (1 or 0)

${BLUE}Notes:${NC}
  - Integration tests create and delete real resources in your Hetzner Cloud account
  - Use a test account, not your production account
  - Ensure you have sufficient resources/quota
  - Tests are skipped if HCLOUD_TOKEN is not set

EOF
}

check_token() {
    if [ -z "$HCLOUD_TOKEN" ]; then
        echo -e "${RED}✗ HCLOUD_TOKEN is not set${NC}"
        echo ""
        echo "To run integration tests, you need to set your Hetzner Cloud API token:"
        echo ""
        echo "  1. Go to https://console.hetzner.cloud"
        echo "  2. Select your project"
        echo "  3. Navigate to Security → Tokens"
        echo "  4. Generate a new token with Read & Write permissions"
        echo "  5. Copy the token and run:"
        echo ""
        echo "    export HCLOUD_TOKEN='your-token-here'"
        echo ""
        return 1
    else
        # Show masked token for verification
        TOKEN_PREFIX=$(echo $HCLOUD_TOKEN | cut -c1-10)
        echo -e "${GREEN}✓ HCLOUD_TOKEN is set${NC}"
        echo "  Token starts with: ${TOKEN_PREFIX}..."
        return 0
    fi
}

run_all_tests() {
    local timeout=300
    local parallel=1
    local verbose=""
    local race=""

    # Parse options
    while [ $# -gt 0 ]; do
        case $1 in
            --timeout)
                timeout=$2
                shift 2
                ;;
            --parallel)
                parallel=$2
                shift 2
                ;;
            --verbose)
                verbose="-v"
                shift
                ;;
            --race)
                race="-race"
                shift
                ;;
            *)
                echo "Unknown option: $1"
                exit 1
                ;;
        esac
    done

    if ! check_token; then
        exit 1
    fi

    echo ""
    echo -e "${BLUE}Running integration tests...${NC}"
    echo "  Timeout: ${timeout}s"
    echo "  Parallel: ${parallel}"
    echo ""

    go test $verbose -tags=integration -timeout ${timeout}s -p $parallel $race ./pkg/hcloud
}

run_single_test() {
    local test_name=$1
    local timeout=300

    if [ -z "$test_name" ]; then
        echo -e "${RED}Error: Test name not provided${NC}"
        echo ""
        echo "Available tests:"
        echo "  - TestHCloudIntegration_CreateAndGetNetwork"
        echo "  - TestHCloudIntegration_UpdateNetwork"
        echo "  - TestHCloudIntegration_ListNetworks"
        echo "  - TestHCloudIntegration_FullLifecycle"
        echo ""
        echo "Usage: $0 run-single <test-name>"
        exit 1
    fi

    if ! check_token; then
        exit 1
    fi

    echo ""
    echo -e "${BLUE}Running integration test: ${test_name}${NC}"
    echo ""

    go test -v -tags=integration -timeout ${timeout}s -run $test_name ./pkg/hcloud
}

# Main
if [ $# -eq 0 ]; then
    print_help
    exit 0
fi

case $1 in
    run)
        shift
        run_all_tests "$@"
        ;;
    run-single)
        shift
        run_single_test "$@"
        ;;
    check-token)
        check_token
        ;;
    help|-h|--help)
        print_help
        ;;
    *)
        echo -e "${RED}Unknown command: $1${NC}"
        echo ""
        print_help
        exit 1
        ;;
esac
