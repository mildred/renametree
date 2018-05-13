A tool to detect renames, specification
=======================================

Goal
----

- Can be used in conjunction with rsync. Executed before rsync to renames files to their new location and rsync can synchronize the rest as it works well for that
- Does detect the file renames that happened in a tree and apply the changes in the other tree

Scenarios
---------

No change:

- Given two directories A and B with similar files in both
- When `renametree A B` is executed
- Then no change is performed
- And the index file is updated for A and B

Simple rename:

- Given two directories A and B already synchronized
- And some files renamed in A
- When I execute `renametree A B`
- Then Files in B are renamed the same way they were renamed in A
- And index for A and B is updated to track the new location of those files

Reverse rename:

- Given two directories A and B already synchronized
- And some files renamed in B
- When I execute `renametree A B`
- Then Files in A are renamed the same way they were renamed in B
- And index for A and B is updated to track the new location of those files

Conflicting rename:

- Given two directories A and B already synchronized
- And some files renamed in A
- And the same files renamed in B with a different name
- When I execute `renametree A B`
- Then the command stops and shows the files that are in rename conflict
- And index for A and B is updated to track the new location of those files

Notes
-----

- A file is a regular file or a directory with an inode. Renames are detected by looking at the inode and finding it in a new location
- A directory tracked by this tool contains an index file which contains the location of each file given their inode
- A rename is detected when the inode is found in another location than specified in the index file
- The index track all earlier locationf of the files too
- The index is updated each time the tool is executed
- Files are associated unique identifiers that are shared across synchronized directories

Algorithm
---------

main:

- record the start time
- for each file in A and B:
    - if the file does not have an id
        - generate an id based on hashing: the start time, the relative path and a hash of the data
    - if the file does 
- we know all files in A and B have an id and their location is recorded
- for each file id that exists in both A and B:
    - if their relative path is the same
        - do nothing, files are already at their correct location
    - if the location history tells one file is on an older location
        - move it to the new location and record the new location history in the index
    - if both files have different locations but history cannot tell which one is the correct location
        - record file to show a conflict error
- show all conflict errors
- exit with status >0 in case of errors

tell history for files A and B

- while each file history is starting with the same location
    - remove each directory from the beginning of each file history
- if a file has an empty history
    - it is the oldest file
    - return other file remaining history
- if both files still have history
    - this is a conflict

Index format
------------

The index file name should contain in its name the inode number of the directory it refers to, so in case the file is copied by rsync elsewhere, it is not read mistakenly.

The index should contain entries with:

- the list of past locations of a file
- the inode number of a file
- the unique id of a file

### Requirements

- given a file name or an opened file, tell a unique id for it across directories
- given a unique file id, tell the history of the locations of this file

Alternative without index
-------------------------

Without an index file synchronized on both parties, a file can be identified by:

- an inode number, but only for the scope of a directory
- a data hash but only at some point in time
- a file name until this is renamed too

It seems impossible to track renames without an index