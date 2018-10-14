#!/bin/bash
# Script to create a software update that will be deployed to CryptoFaxPA
set -euo pipefail

# The checkout must be pristine
if output=$(git status --porcelain) && [ ! -z "$output" ]; then
	echo "release.sh must be run only a pristine checkout."
	echo
	echo "Unexpected contents:"
	git status --porcelain
	exit 1
fi

APPS="client wificonf"

# Build the binaries
for app in $APPS; do
	GOOS=linux GOARCH=arm GOARM=7 \
		go build -ldflags="-s -w" -o "$(pwd)/overlay/home/pi/$app" "./$app"
done

VHASH=$(git log --pretty=format:'%h' -n1)
VDATE=$(git log --pretty=format:'%ci' -n1)
KEYRING="./overlay/home/pi/trusted.gpg"   # MUST use explicitly-relative path
ASSET_SWFAX="firmware.swfax"
GHAPI="https://api.github.com/repos/rasky/cryptofaxpa"

# Temporary files
TMP_SWFAX=$(mktemp "${TMPDIR:-/tmp/}$(basename "$0").XXXXXXXXXXXX")
TMP_CURL=$(mktemp "${TMPDIR:-/tmp/}$(basename "$0").XXXXXXXXXXXX")

trap 'rm -f "$TMP_SWFAX" "$TMP_CURL"' EXIT

# Package release
echo "Creating release..."
case $(uname -s) in
    Darwin*)    TAR=gtar;;
    *)          TAR=tar
esac
$TAR --mode=a+rw --mtime="$VDATE" --owner=0 --group=0 \
	-C overlay --exclude=.KEEPME -cz . | gpg --quiet --sign > "$TMP_SWFAX"

# Verify the release is made with a trusted version
if ! gpg --no-default-keyring --no-auto-key-retrieve --keyring "$KEYRING" --verify "$TMP_SWFAX"; then
	echo
	echo "FATAL: Cannot gpg-sign the release with a key which is not trusted"
	echo "(not present in $KEYRING)"
	exit 1
fi

# Latest GitHub release ID
RELID=$(curl -s $GHAPI/releases/latest | jq -r .id)
if [ "$RELID" == "" ]; then
	echo "No release found on GitHub in the cryptofax repo (?)"
	exit 1
fi

# See if there's already a firmware there, and if so delete it
ASSETID=$(curl -s "$GHAPI/releases/$RELID/assets" | jq -r ".[] | select(.name==\"$ASSET_SWFAX\") | .id")

# Upload the release to GitHub
echo -n "Enter your GitHub username: "
read -r GH_USERNAME

echo -n "Enter your GitHub password: "
read -r -s GH_PASSWORD
echo

echo -n "Enter your GitHub 2FA OTP: "
read -r GH_2FA

if [ "$ASSETID" != "" ]; then
	echo "Removing previous asset from GitHub..."
	HTTP_RES=$(curl -s -XDELETE \
		-u "$GH_USERNAME:$GH_PASSWORD" -H "X-GitHub-OTP: $GH_2FA" \
		--write-out "%{http_code}" \
		-o "$TMP_CURL" \
		"$GHAPI/releases/assets/$ASSETID")
	if [ "${HTTP_RES::1}" != "2" ]; then
		echo "Deletion failed: HTTP code = $HTTP_RES"
		cat "$TMP_CURL"
		exit 1
	fi
fi

echo "Uploading release asset to GitHub..."
HTTP_RES=$(curl -# -XPOST -H "Content-Type:application/octet-stream" \
	-u "$GH_USERNAME:$GH_PASSWORD" -H "X-GitHub-OTP: $GH_2FA" \
	--data-binary "@$TMP_SWFAX" \
	--write-out "%{http_code}" \
	-o "$TMP_CURL" \
	"https://uploads.github.com/repos/rasky/cryptofaxpa/releases/$RELID/assets?name=$ASSET_SWFAX&label=Firmware%20built%20at%20$VHASH")
if [ "${HTTP_RES::1}" != "2" ]; then
	echo "Upload failed: HTTP code = $HTTP_RES"
	cat "$TMP_CURL"
	exit 1
fi

echo "Release uploaded successfully"
echo "See the release here: https://github.com/rasky/cryptofaxpa/releases/latest"
