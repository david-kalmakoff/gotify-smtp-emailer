 #!/bin/sh

cat ci/SUPPORTED_VERSIONS.txt | while read TARGET; do
	make GOTIFY_VERSION="$TARGET" FILE_SUFFIX="-for-gotify-$TARGET" build || exit 1
done
