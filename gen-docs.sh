#!/bin/bash

USAGE="$(./twitter-follower-rank -help 2>&1 | gsed 's/\t/    /g' | gsed ':a;N;$!ba;s/\n/\\n/g' | gsed -E "s/\(/\\\(/g" | gsed -E "s/\)/\\\)/g")"
perl -i -pe 'BEGIN{undef $/;} s/(<!--- GENERATED FROM MAKEFILE -->).*?(<!--- GENERATED FROM MAKEFILE -->)/$1\n```bash\nTWITTER_FOLLOWER_RANK_HELP\n```\n$2/smg' README.md
gsed -i "s|TWITTER_FOLLOWER_RANK_HELP|${USAGE}|g" README.md
