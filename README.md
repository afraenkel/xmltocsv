# xmltocsv

xmltocsv is a tool to convert a list of (perhaps nested) xml blobs to a flattened list of records in csv format. Assumes that each XML blob are single line.

## To Do

* command line flags: usage messages / split out into init
* restructure into packages instead of a single file
* more tests (multiline xml, different delim, keysep, API tests)
* fix indexing of repeated keys to 1,2,3,4... from *,1,2,3...
* play with concurrency