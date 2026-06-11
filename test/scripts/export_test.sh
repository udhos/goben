#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
BIN_DIR="/tmp/goben_test_bins"
OUTPUT_DIR="$PROJECT_DIR/test/output"

cleanup() {
    rm -rf "$BIN_DIR"
    pkill -f "goben_test_server" 2>/dev/null || true
}
trap cleanup EXIT

build() {
    echo "[build] compiling binaries..."
    mkdir -p "$BIN_DIR" "$OUTPUT_DIR"
    go build -o "$BIN_DIR/goben_test_client" "$PROJECT_DIR/cmd/goben"
    cp "$BIN_DIR/goben_test_client" "$BIN_DIR/goben_test_server"
}

wait_for_server() {
    local port=$1
    local max=10
    local i
    for ((i = 0; i < max; i++)); do
        if ss -tlnp 2>/dev/null | grep -q ":${port} "; then
            return 0
        fi
        sleep 0.5
    done
    echo "[FAIL] server on port $port did not start"
    return 1
}

run_server() {
    local port=$1
    "$BIN_DIR/goben_test_server" \
        -l "127.0.0.1:$port" \
        -s=false -t=true -u=false &
    wait_for_server "$port"
}

assert_files_exist() {
    for f in "$@"; do
        local path="$OUTPUT_DIR/$f"
        if [ ! -f "$path" ]; then
            echo "[FAIL] file not found: $path"
            return 1
        fi
        echo "[OK]   $f ($(stat -c%s "$path") bytes)"
    done
}

assert_files_not_exist() {
    for f in "$@"; do
        local path="$OUTPUT_DIR/$f"
        if [ -f "$path" ]; then
            echo "[FAIL] file should not exist: $path"
            return 1
        fi
    done
}

assert_no_malformed_files() {
    local count
    count=$(find "$OUTPUT_DIR" -maxdepth 1 -name '*!\(*' 2>/dev/null | wc -l)
    if [ "$count" -gt 0 ]; then
        echo "[FAIL] found files with malformed names:"
        find "$OUTPUT_DIR" -maxdepth 1 -name '*!\(*'
        return 1
    fi
}

run_client() {
    (cd "$OUTPUT_DIR" && "$BIN_DIR/goben_test_client" "$@")
}

test_default_ascii_console_only() {
    echo ""
    echo "=== Test: default behavior (no --export) outputs ASCII to console only ==="
    local port=19001

    run_server "$port"
    run_client \
        -H "127.0.0.1:$port" \
        -s=false -t=true -u=false \
        -d 3s 2>&1 | grep -E "input:|output:" || true

    assert_files_not_exist "result-0-127.0.0.1:$port.ascii"
    echo "[OK]   no file written for default ASCII (console only)"
    pkill -f "goben_test_server" 2>/dev/null || true
    sleep 0.5
}

test_explicit_ascii_writes_file() {
    echo ""
    echo "=== Test: --export ascii writes to console AND file ==="
    local port=19002

    run_server "$port"
    run_client \
        -H "127.0.0.1:$port" \
        -s=false -t=true -u=false \
        -d 3s \
        --export ascii 2>&1 | grep -E "input:|output:" || true

    assert_files_exist "result-0-127.0.0.1:$port.ascii"
    pkill -f "goben_test_server" 2>/dev/null || true
    sleep 0.5
}

test_comma_separated_modes() {
    echo ""
    echo "=== Test: -e csv,yaml,png generates all three files ==="
    local port=19003

    run_server "$port"
    run_client \
        -H "127.0.0.1:$port" \
        -s=false -t=true -u=false \
        -d 3s \
        -e csv,yaml,png 2>&1 | grep -E "export|render" || true

    assert_files_exist \
        "result-0-127.0.0.1:$port.csv" \
        "result-0-127.0.0.1:$port.yaml" \
        "result-0-127.0.0.1:$port.png"
    assert_files_not_exist "result-0-127.0.0.1:$port.ascii"

    pkill -f "goben_test_server" 2>/dev/null || true
    sleep 0.5
}

test_repeated_flags() {
    echo ""
    echo "=== Test: -e X.yaml -e X.png -e X.csv with fixed filenames ==="
    local port=19004

    run_server "$port"
    run_client \
        -H "127.0.0.1:$port" \
        -s=false -t=true -u=false \
        -d 3s \
        -e mytest.yaml -e mytest.png -e mytest.csv 2>&1 | grep -E "export|render" || true

    assert_files_exist mytest.yaml mytest.png mytest.csv
    assert_no_malformed_files

    pkill -f "goben_test_server" 2>/dev/null || true
    sleep 0.5
}

test_all_formats() {
    echo ""
    echo "=== Test: -e csv,yaml,png,ascii generates all four files ==="
    local port=19005

    run_server "$port"
    run_client \
        -H "127.0.0.1:$port" \
        -s=false -t=true -u=false \
        -d 3s \
        -e csv,yaml,png,ascii 2>&1 | grep -E "export|render" || true

    assert_files_exist \
        "result-0-127.0.0.1:$port.csv" \
        "result-0-127.0.0.1:$port.yaml" \
        "result-0-127.0.0.1:$port.png" \
        "result-0-127.0.0.1:$port.ascii"

    pkill -f "goben_test_server" 2>/dev/null || true
    sleep 0.5
}

test_mixed_modes_and_filenames() {
    echo ""
    echo "=== Test: -e ascii,result-%d-%s.csv,myreport.png mixed formats ==="
    local port=19006

    run_server "$port"
    run_client \
        -H "127.0.0.1:$port" \
        -s=false -t=true -u=false \
        -d 3s \
        -e "ascii,result-%d-%s.csv,myreport.png" 2>&1 | grep -E "export|render" || true

    assert_files_exist \
        "result-0-127.0.0.1:$port.ascii" \
        "result-0-127.0.0.1:$port.csv" \
        "myreport.png"

    pkill -f "goben_test_server" 2>/dev/null || true
    sleep 0.5
}

main() {
    cd "$PROJECT_DIR"
    build

    PASSED=0
    FAILED=0

    for test_fn in \
        test_default_ascii_console_only \
        test_explicit_ascii_writes_file \
        test_comma_separated_modes \
        test_repeated_flags \
        test_all_formats \
        test_mixed_modes_and_filenames; do
        if $test_fn; then
            PASSED=$((PASSED + 1))
        else
            FAILED=$((FAILED + 1))
        fi
    done

    echo ""
    echo "==============================="
    echo "  PASSED: $PASSED  FAILED: $FAILED"
    echo "==============================="

    if [ "$FAILED" -gt 0 ]; then
        exit 1
    fi
}

main "$@"
