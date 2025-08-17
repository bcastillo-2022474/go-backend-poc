# GoLand Protobuf Import Resolution Issue

## Problem
GoLand shows "Cannot resolve import" error for `google/api/annotations.proto` but not for `google/protobuf/timestamp.proto`, even though `buf generate` works perfectly.

## Root Cause
GoLand has **built-in protobuf definitions** bundled in its plugin:
- ✅ `google/protobuf/*` → Built into GoLand at `/home/user/.local/share/JetBrains/Toolbox/apps/goland/plugins/protoeditor/lib/protoeditor.jar!/include/`
- ❌ `google/api/*` → Not bundled, requires manual configuration

## Solutions

### Option 1: Configure GoLand Import Paths
1. Go to **File → Settings → Languages & Frameworks → Protocol Buffers**
2. Click **+** to add import paths
3. Add buf cache path: `~/.cache/buf/v1/module/data/buf.build/googleapis/googleapis/[COMMIT_HASH]`

**Find your commit hash:**
```bash
# Look in buf.lock for googleapis commit
cat buf.lock | grep -A1 "googleapis/googleapis"
```

### Option 2: Ignore the Warning (Recommended)
Since `buf generate` works perfectly and code compiles fine, this is purely cosmetic. Many developers live with this red squiggle.

## Verification
- ✅ `buf generate` works without errors
- ✅ Generated code compiles successfully  
- ✅ All protobuf dependencies resolved at build time
- ❌ IDE shows false positive import error

## Note
This is a common issue with protobuf IDEs and external dependency managers like buf. The actual compilation and code generation work correctly regardless of the IDE warning.