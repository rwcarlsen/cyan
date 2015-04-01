#!/bin/bash

sql=$2
case $2 in
"flow-proto")
    sql="SELECT tl.Time,IFNULL(sub.qty,0) FROM timelist as tl
         LEFT JOIN (
            SELECT t.time as time,SUM(r.quantity) as qty FROM transactions as t
            JOIN resources as r on t.resourceid=r.resourceid
            JOIN agents as send ON t.senderid=send.agentid
            JOIN agents as recv ON t.receiverid=recv.agentid
            WHERE send.prototype='$3' AND recv.prototype='$4'
            GROUP BY t.time
         ) AS sub ON tl.time=sub.time;
        "
;;
"inv-proto")
    sql="SELECT tl.Time,IFNULL(sub.qty,0) FROM timelist as tl
         LEFT JOIN (SELECT tl.Time as time,TOTAL(inv.Quantity) AS qty FROM timelist as tl
            JOIN inventories as inv on inv.starttime <= tl.time and inv.endtime > tl.time
            JOIN agents as a on a.agentid=inv.agentid
            WHERE a.prototype='$3'
            GROUP BY tl.Time
         ) AS sub ON sub.time=tl.time;
         "
;;
"inv-proto-nuc")
    sql="SELECT tl.Time,IFNULL(sub.qty,0) FROM timelist as tl
         LEFT JOIN (SELECT tl.Time as time,TOTAL(inv.Quantity) AS qty FROM timelist as tl
            JOIN inventories as inv on inv.starttime <= tl.time and inv.endtime > tl.time
            JOIN compositions as c on c.qualid=inv.qualid
            JOIN agents as a on a.agentid=inv.agentid
            WHERE a.prototype='$3' AND c.NucId IN ($4)
            GROUP BY tl.Time
         ) AS sub ON sub.time=tl.time;
         "
;;
esac

postpath=$(which cycpost)
if [[ -n $postpath ]]; then
    cycpost $1 > /dev/null
fi

sqlite3 -column "$1" "$sql"
#fname=.$(uuidgen).dat
#sqlite3 -column "$1" "$sql" > $fname
#gnuplot -p -e "plot "\""$fname"\"" using 1:2 with linespoints;"
#rm $fname

