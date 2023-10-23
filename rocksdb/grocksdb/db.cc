// https://techoverflow.net/2020/01/28/rocksdb-minimal-example-in-c/
#include <cassert>
#include <string>
#include <rocksdb/db.h>
#include <rocksdb/slice_transform.h>
using namespace std;
using namespace rocksdb;
int main(int argc, char** argv) {
    DB* db;
    Options options;
    options.create_if_missing = true;
    options.prefix_extractor.reset(NewNoopTransform());

    Status status =
    DB::Open(options, "/tmp/testdb", &db);
    assert(status.ok());
    db->Close();
}