pkill KeyValue
for i in {1..7}
do
leveldb_check leveldb2_$i.db/| wc -l
done

