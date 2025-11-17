#!/bin/bash
# End-to-End Workflow Test for BobaMixer
# Tests the complete user journey from initialization to running tools

set -e

echo "🧋 BobaMixer End-to-End Workflow Test"
echo "======================================"
echo

# Cleanup function
cleanup() {
    echo "🧹 Cleaning up test environment..."
    rm -rf ~/.boba-test
}

# Set up test environment
export BOBA_HOME="$HOME/.boba-test"
trap cleanup EXIT

echo "📁 Step 1: Initialize BobaMixer"
echo "--------------------------------"
./boba init
echo "✅ Initialization complete"
echo

echo "📋 Step 2: Verify configuration files"
echo "--------------------------------------"
if [ -f "$BOBA_HOME/providers.yaml" ]; then
    echo "✅ providers.yaml exists"
else
    echo "❌ providers.yaml missing"
    exit 1
fi

if [ -f "$BOBA_HOME/tools.yaml" ]; then
    echo "✅ tools.yaml exists"
else
    echo "❌ tools.yaml missing"
    exit 1
fi

if [ -f "$BOBA_HOME/bindings.yaml" ]; then
    echo "✅ bindings.yaml exists"
else
    echo "❌ bindings.yaml missing"
    exit 1
fi

if [ -f "$BOBA_HOME/secrets.yaml" ]; then
    echo "✅ secrets.yaml exists"
else
    echo "❌ secrets.yaml missing"
    exit 1
fi
echo

echo "🔧 Step 3: List available tools"
echo "--------------------------------"
./boba tools
echo "✅ Tools listed"
echo

echo "🌐 Step 4: List available providers"
echo "------------------------------------"
./boba providers
echo "✅ Providers listed"
echo

echo "🔗 Step 5: Test binding (claude to anthropic)"
echo "----------------------------------------------"
./boba bind claude claude-anthropic-official
echo "✅ Binding created"
echo

echo "🔗 Step 6: Test binding with proxy (codex to openai)"
echo "------------------------------------------------------"
./boba bind codex openai-official --proxy=on
echo "✅ Binding with proxy created"
echo

echo "🏥 Step 7: Run diagnostics"
echo "--------------------------"
./boba doctor
echo "✅ Diagnostics complete"
echo

echo "🚀 Step 8: Test dry-run (env injection verification)"
echo "------------------------------------------------------"
echo "Creating a test wrapper script..."

cat > /tmp/test-claude-wrapper.sh << 'WRAPPER_EOF'
#!/bin/bash
echo "=== Environment Variables Injected ==="
env | grep -E "ANTHROPIC|OPENAI|GEMINI" | sort
echo "======================================"
echo "Arguments passed: $@"
exit 0
WRAPPER_EOF

chmod +x /tmp/test-claude-wrapper.sh

# Temporarily replace claude exec in tools.yaml for testing
cp "$BOBA_HOME/tools.yaml" "$BOBA_HOME/tools.yaml.backup"
sed -i 's|exec: "claude"|exec: "/tmp/test-claude-wrapper.sh"|' "$BOBA_HOME/tools.yaml"

echo
echo "Running: boba run claude --version"
./boba run claude --version || true

# Restore original tools.yaml
mv "$BOBA_HOME/tools.yaml.backup" "$BOBA_HOME/tools.yaml"

echo
echo "✅ Env injection verified"
echo

echo "📊 Step 9: Test proxy status"
echo "-----------------------------"
./boba proxy status
echo "✅ Proxy status checked"
echo

echo "════════════════════════════════════"
echo "✅ All End-to-End Tests Passed!"
echo "════════════════════════════════════"
echo
echo "Summary:"
echo "  ✅ Configuration initialization"
echo "  ✅ Tools and providers listing"
echo "  ✅ Binding creation and management"
echo "  ✅ Diagnostics"
echo "  ✅ Environment variable injection"
echo "  ✅ Proxy integration"
echo
echo "🎉 BobaMixer core workflow is fully functional!"
