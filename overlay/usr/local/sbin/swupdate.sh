#!/bin/bash
set -euo pipefail

GHAPI="https://api.github.com/repos/rasky/cryptofaxpa"
ASSET_SWFAX="firmware.swfax"
KEYRING="/home/pi/trusted.gpg"
LAST_TIMESTAMP="/var/cache/firmware.last_updated"

# Latest GitHub release ID
RELID=$(curl -s $GHAPI/releases/latest | jq -r .id)
if [ "$RELID" == "" ]; then
	echo "No release found on GitHub in the cryptofax repo (?)"
	exit 1
fi

ASSET=$(curl -s "$GHAPI/releases/$RELID/assets" | jq -r ".[] | select(.name==\"$ASSET_SWFAX\")")

UPDATED_AT=$(echo "$ASSET" | jq -r .updated_at)
URL=$(echo "$ASSET" | jq -r .browser_download_url)

touch $LAST_TIMESTAMP   # make sure it exists, even if empty
if [ "$(cat $LAST_TIMESTAMP)" == "$UPDATED_AT" ]; then
	echo "No new firmware found (last updated: $UPDATED_AT)"
	exit 0
fi

trap 'rm -f "/tmp/$ASSET_SWFAX" "/tmp/$ASSET_SWFAX.tgz"' EXIT

echo "Downloading new firmware: $URL"
curl -L -s "$URL" >/tmp/$ASSET_SWFAX

echo "Verifying new firmware"
if ! gpg --no-default-keyring --no-auto-key-retrieve \
	--keyring "$KEYRING" \
	--decrypt "/tmp/$ASSET_SWFAX" >"/tmp/$ASSET_SWFAX.tgz"; then
	echo "FAILED: gpg signature failure"
	exit 1
fi

# Verify the archive before extracting, just in case
tar -tf "/tmp/$ASSET_SWFAX.tgz"

# Extract the new software. Use unlink-first to ovewrite running binaries
echo "Extracting firmware..."
tar -C / -xf "/tmp/$ASSET_SWFAX.tgz"

# Update timestamp
echo "$UPDATED_AT" >$LAST_TIMESTAMP

echo "Firmware updated successfully!"
echo "Rebooting"
reboot
