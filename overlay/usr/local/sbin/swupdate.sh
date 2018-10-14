#!/bin/bash
set -euo pipefail

echo "Starting software update..."

GHAPI="https://api.github.com/repos/rasky/cryptofaxpa"
ASSET_SWFAX="firmware.swfax"
KEYRING="/home/pi/trusted.gpg"
LAST_TIMESTAMP="/var/cache/firmware.last_updated"

# Do a http_get request to the specified URL.
# When the HTTP result code is 2xx, outputs the body of the result on stdout.
# When the HTTP result code is not 2xx (or the request fails at the HTTP-level
# for other errors), an error message that includes the response body is
# output to stderr, and the script is aborted.
function http_get() {
	TMP_CURL=$(mktemp "${TMPDIR:-/tmp/}$(basename "$0").XXXXXXXXXXXX")
	trap 'rm -f "$TMP_CURL"' EXIT
	HTTP_RES=$(curl -L -s \
		--write-out "%{http_code}" \
		-o "$TMP_CURL" \
		"$1")
	if [ "${HTTP_RES::1}" != "2" ]; then
		>&2 echo "HTTP request failed: code = $HTTP_RES"
		>&2 cat "$TMP_CURL"
		exit 1
	fi
	cat "$TMP_CURL"
	rm -f "$TMP_CURL"
}

# Fetch the latest GitHub release ID
RELID=$(http_get $GHAPI/releases/latest | jq -r .id)
if [ "$RELID" == "" ]; then
	echo "No release found on GitHub in the cryptofax repo (?)"
	exit 1
fi

# Fetch the JSON description of the asset whose name is $ASSET_SWFAX.
ASSET=$(http_get "$GHAPI/releases/$RELID/assets" | jq -r ".[] | select(.name==\"$ASSET_SWFAX\")")

# Get the last-udpated timestamp and direct download URL of the asset
UPDATED_AT=$(echo "$ASSET" | jq -r .updated_at)
URL=$(echo "$ASSET" | jq -r .browser_download_url)

# If the timestamp has not changed since the last update, exit without error
touch $LAST_TIMESTAMP   # make sure it exists, even if empty
if [ "$(cat $LAST_TIMESTAMP)" == "$UPDATED_AT" ]; then
	echo "No new firmware found (last updated: $UPDATED_AT)"
	exit 0
fi

trap 'rm -f "/tmp/$ASSET_SWFAX" "/tmp/$ASSET_SWFAX.tgz"' EXIT

echo "Downloading new firmware: $URL"
http_get "$URL" >/tmp/$ASSET_SWFAX

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

# SPECIAL CASE for wpa_supplicant. Update only if not already updated. We use
# "update_config=1" has key to check, as it's the most important one.
if ! grep -q ^update_config=1 /etc/wpa_supplicant/wpa_supplicant.conf; then
	mv /etc/wpa_supplicant/wpa_supplicant.template.conf /etc/wpa_supplicant/wpa_supplicant.conf
fi

echo "Firmware updated successfully!"
echo "Rebooting"
reboot
