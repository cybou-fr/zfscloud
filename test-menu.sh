#!/bin/bash
# Test runner - Menu-driven interface for all test scripts

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m'

show_banner() {
    clear
    echo -e "${MAGENTA}"
    echo "╔═══════════════════════════════════════════════════════╗"
    echo "║                                                       ║"
    echo "║         ZFSCloud Testing Suite for Linux             ║"
    echo "║                                                       ║"
    echo "╚═══════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

show_menu() {
    echo ""
    echo -e "${CYAN}Available Test Options:${NC}"
    echo ""
    echo -e "${GREEN}1)${NC} Quick Test          - Start server with manual test instructions"
    echo -e "${GREEN}2)${NC} Integration Test   - Automated full workflow test"
    echo -e "${GREEN}3)${NC} Performance Test   - Benchmark upload/download speeds"
    echo -e "${GREEN}4)${NC} Custom Test        - Interactive helper utilities"
    echo ""
    echo -e "${YELLOW}5)${NC} Build Binaries     - Rebuild server and client"
    echo -e "${YELLOW}6)${NC} System Info        - Show server and ZFS status"
    echo -e "${YELLOW}7)${NC} Server Logs        - View server container logs"
    echo ""
    echo -e "${RED}8)${NC} Stop Server        - Stop Docker container"
    echo -e "${RED}9)${NC} Clean All          - Remove all data and containers"
    echo ""
    echo -e "${BLUE}0)${NC} Exit"
    echo ""
}

press_any_key() {
    echo ""
    echo -e "${YELLOW}Press any key to continue...${NC}"
    read -n 1 -s
}

run_test() {
    local test_name="$1"
    local test_script="$2"
    
    echo -e "${CYAN}Running $test_name...${NC}"
    echo ""
    
    if [ -f "$test_script" ]; then
        bash "$test_script"
    else
        echo -e "${RED}Error: $test_script not found${NC}"
    fi
    
    press_any_key
}

main() {
    while true; do
        show_banner
        
        # Check if server is running
        if docker ps | grep -q zfscloud-server; then
            echo -e "${GREEN}● Server Status: Running${NC}"
        else
            echo -e "${YELLOW}● Server Status: Stopped${NC}"
        fi
        
        # Check if binaries exist
        if [ -f "bin/zfscloud-server" ] && [ -f "bin/zfscloud-client" ]; then
            echo -e "${GREEN}● Binaries: Built${NC}"
        else
            echo -e "${YELLOW}● Binaries: Not found${NC}"
        fi
        
        show_menu
        
        echo -n "Select option: "
        read -r choice
        
        case $choice in
            1)
                run_test "Quick Test" "./test-docker.sh"
                ;;
            2)
                run_test "Integration Test" "./test-integration.sh"
                ;;
            3)
                echo ""
                echo -e "${CYAN}Performance Test Configuration:${NC}"
                echo -n "Number of clients (default 3): "
                read -r num_clients
                echo -n "Large file size in MB (default 50): "
                read -r file_size
                
                export NUM_CLIENTS=${num_clients:-3}
                export FILE_SIZE_LARGE=${file_size:-50}
                
                run_test "Performance Test" "./test-performance.sh"
                ;;
            4)
                echo ""
                echo -e "${CYAN}Helper Utilities:${NC}"
                ./test-helpers.sh
                echo ""
                echo -e "${YELLOW}Example: ./test-helpers.sh show_system_info${NC}"
                echo -e "${YELLOW}Example: ./test-helpers.sh db_list_users${NC}"
                press_any_key
                ;;
            5)
                echo ""
                echo -e "${CYAN}Building binaries...${NC}"
                ./test-helpers.sh build_all
                press_any_key
                ;;
            6)
                echo ""
                ./test-helpers.sh show_system_info
                press_any_key
                ;;
            7)
                echo ""
                echo -e "${CYAN}Server Logs (last 50 lines):${NC}"
                docker logs --tail=50 zfscloud-server 2>/dev/null || echo -e "${RED}Server not running${NC}"
                echo ""
                echo -e "${YELLOW}Use 'docker logs -f zfscloud-server' to follow logs${NC}"
                press_any_key
                ;;
            8)
                echo ""
                echo -e "${YELLOW}Stopping server...${NC}"
                cd docker
                docker-compose down
                cd ..
                echo -e "${GREEN}Server stopped${NC}"
                press_any_key
                ;;
            9)
                echo ""
                echo -e "${RED}WARNING: This will delete ALL data!${NC}"
                echo -n "Are you sure? (yes/no): "
                read -r confirm
                if [ "$confirm" = "yes" ]; then
                    echo -e "${YELLOW}Cleaning all data...${NC}"
                    cd docker
                    docker-compose down -v
                    cd ..
                    echo -e "${GREEN}All data cleaned${NC}"
                else
                    echo "Cancelled"
                fi
                press_any_key
                ;;
            0)
                echo ""
                echo -e "${GREEN}Goodbye!${NC}"
                exit 0
                ;;
            *)
                echo ""
                echo -e "${RED}Invalid option. Please try again.${NC}"
                sleep 2
                ;;
        esac
    done
}

# Run main menu
main
