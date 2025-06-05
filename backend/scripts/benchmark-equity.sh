#!/bin/bash

# Equity calculation benchmark script

echo "=== Equity Calculation Benchmark ==="
echo ""

# テスト用の日付（実際にはDBに保存しないが、計算は実行される）
TEST_DATE="2099-12-31"

# Build the batch program
echo "Building batch program..."
cd /Users/shinichiro/Products/equity-distribution-app/backend
go build -o test-batch ./batch/main.go

# Run exhaustive version (1 scenario only for quick test)
echo ""
echo "Running EXHAUSTIVE version (processing first scenario only)..."
echo "Start time: $(date)"
START_TIME=$(date +%s)

timeout 300 ./test-batch -date $TEST_DATE -monte-carlo=false -image-upload=false -parallel=false 2>&1 | head -20

END_TIME=$(date +%s)
EXHAUSTIVE_TIME=$((END_TIME - START_TIME))
echo "End time: $(date)"
echo "Exhaustive calculation took: ${EXHAUSTIVE_TIME} seconds"

# Run Monte Carlo version
echo ""
echo "Running MONTE CARLO version (processing first scenario only)..."
echo "Start time: $(date)"
START_TIME=$(date +%s)

timeout 300 ./test-batch -date $TEST_DATE -monte-carlo=true -image-upload=false -parallel=false 2>&1 | head -20

END_TIME=$(date +%s)
MONTE_CARLO_TIME=$((END_TIME - START_TIME))
echo "End time: $(date)"
echo "Monte Carlo calculation took: ${MONTE_CARLO_TIME} seconds"

# Summary
echo ""
echo "=== SUMMARY ==="
echo "Exhaustive: ${EXHAUSTIVE_TIME} seconds"
echo "Monte Carlo: ${MONTE_CARLO_TIME} seconds"

if [ $EXHAUSTIVE_TIME -gt 0 ] && [ $MONTE_CARLO_TIME -gt 0 ]; then
    SPEEDUP=$(echo "scale=2; $EXHAUSTIVE_TIME / $MONTE_CARLO_TIME" | bc)
    echo "Speedup: ${SPEEDUP}x"
fi

# Cleanup
rm -f test-batch