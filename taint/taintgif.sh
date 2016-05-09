#!/bin/bash

db=$1
simdur=$2
res=$3

rm -f taint-frame-*.gif

for t in $(seq $simdur); do
    echo "writing frame ${t}..."
    cyan -db $db taint -res $res -t $t | dot -Tgif > taint-frame-${t}.gif
done

convert $(for a in $(ls -v taint-frame-*.gif); do printf -- "-delay 15 %s " $a; done; ) taint-movie.gif

