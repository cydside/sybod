# sybod

Shrink Your BoltDB: it helps to shrink your BoltDB file to fit the real size.

When you are testing CRUD operation on your BoltDB file and it only grows and never will adapt to the real size even if you delete all data inside, then this tool will help you to copy all the data to a new BoltDB file with the exact size.

This project is based on [bbolt](https://github.com/coreos/bbolt) as the boltdb's author [suggest](https://github.com/boltdb/bolt#a-message-from-the-author).

## Usage

	sybod data.db
  
It will create a copy of data.db named:

	copy_data.db
  
Try it and let me know!
