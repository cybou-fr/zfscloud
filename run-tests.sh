#!/bin/bash
# Test runner - executes all ZFSCloud tests in sequence

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$SCRIPT_DIR"

# Source utilities
source "$SCRIPT_DIR/tests/test-utils.sh"

# Parse arguments
RUN_BASIC=1
RUN_INTEGRATION=0
RUN_PERFORMANCE=0
CLEAN_AFTER=0

while [[ $# -gt 0 ]]; do
    case $1 in
        --all)
            RUN_BASIC=1
            RUN_INTEGRATION=1
            RUN_PERFORMANCE=1
            shift
            ;;
        --integration)
            RUN_INTEGRATION=1
            shift
            ;;
        --performance)
            RUN_PERFORMANCE=1
            shift
            ;;
        --clean)
            CLEAN_AFTER=1
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--all] [--integration] [--performance] [--clean]"
            exit 1
            ;;
    esac
done

log_section "ZFSCloud Test Runner"
echo "Test Suite Selection:"
echo "  Basic Tests: $([ $RUN_BASIC -eq 1 ] && echo 'YES' || echo 'NO')"
echo "  Integration Tests: $([ $RUN_INTEGRATION -eq 1 ] && echo 'YES' || echo 'NO')"
echo "  Performance Tests: $([ $RUN_PERFORMANCE -eq 1 ] && echo 'YES' || echo 'NO')"
echo ""

# Validate prerequisites
log_info "Checking prerequisites..."
if ! validate_prerequisites; then
    log_error "Prerequisites check failed"
    exit 1
fi
log_success "Prerequisites OK"

cd "$PROJECT_DIR"

# Run basic test
if [ $RUN_BASIC -eq 1 ]; then
    log_section "Running Basic Test"
    if [ -f "$PROJECT_DIR/test-docker.sh" ]; then
        "$PROJECT_DIR/test-docker.sh"
    else
        log_error "test-docker.sh not found"
    fi
fi

# Run integration tests
if [ $RUN_INTEGRATION -eq 1 ]; then
    log_section "Running Integration Tests"
    if [ -f "$PROJECT_DIR/tests/integration-test.sh" ]; then
        "$PROJECT_DIR/tests/integration-test.sh"
    else
        log_error "integration-test.sh not found"
    fi
fi

# Run performance tests
if [ $RUN_PERFORMANCE -eq 1 ]; then
    log_section "Running Performance Tests"
    if [ -f "$PROJECT_DIR/tests/performance-test.sh" ]; then
        "$PROJECT_DIR/tests/performance-test.sh"
    else
        log_error "performance-test.sh not found"
    fi
fi

# Cleanup
if [ $CLEAN_AFTER -eq 1 ]; then
    log_section "Cleanup"
    cleanup_docker
    cleanup_test_files "test-*.txt"
    log_success "Cleanup complete"
fi

log_section "Test Runner Complete"
log_success "All requested tests have been executed"
echo ""
echo "To stop the server:"
echo "  cd docker && docker-compose down"
echo ""
