#!/bin/bash

export DISPLAY=:0

process=answer
makerun="/root/answer/answer"

pgrep answer

if ! pgrep $process > /dev/null
then
    echo 'run answer...'
    pkill answer
    rm  answer.log -f
    $makerun -C /data run > answer.log 2>&1 &
else 
   echo `! pgrep $process`
fi



# tail -f answer.log
