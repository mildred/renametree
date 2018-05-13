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
        - if the file exists in another hierarchy
            - reuse the same uuid as in the other hierarchy
        - optionally, if another file exists at that location
            - associate the inode number to that file uuid
        - else
            - optionally, generate an id based on hashing: the start time, the relative path and a hash of the data
    - if the file locations is different than in index
        - record the new file location in the index with the start time
- we know all files in A and B have an id and their location is recorded
- for each file id that exists in both A and B:
    - if their relative path is the same
        - do nothing, files are already at their correct location
    - if the location history tells one file is on an older location
        - move it to the new location and record the new location history in the index
    - if both files have different locations but history cannot tell which one is the correct location
        - record file to show a conflict error
- for each file that has the same path in A and B
    - if the uuid is the same, do nothing
    - else, report a conflict error
- show all conflict errors
- exit with status >0 in case of errors

tell history for files A and B

- starting by the end of each file history, find a location that is the same on both files at the same time
- if not found
    - this is a conflict, return now
- else, process only history from that point in time to the end of history.
    - while the two histories have the same path as first entry in the history
        - remove that entry from both histories
    - if there is a file with an empty history
        - that file is older, return the remaining history of the new file which contains the new location  
    - else, both files still have an history continuing
        - return that there is a conflict

Solving a conflict is as easy as renaming a file in one copy to the same path of the other copy and running the tool again. it will record the same path for both copies of the file for the same point in time and be happy with it.

Solving uuid conflicts is more difficult, it requires a special mode of running
renametree to change the uuid of an existing file.

Index format
------------

The index file name should contain in its name the inode number of the directory it refers to, so in case the file is copied by rsync elsewhere, it is not read mistakenly.

### Requirements

- given a file name or an opened file, tell a unique id for it across directories
- given a file name or an opened file, create a new entry if it does not exists
- given a file name or an opened file, add a directory to the location history, with time
- given a unique file id, tell the timed history of the locations of this file

### Format v0

JSON for simplicity

### Format v1

- The database is named `.renametree-%d` where %d is the inode number of the directory the file is on
- The database is a line based text file
- the first work up to the first TAB or SPACE character tells the line type:
    - `FID <UUID> <INUM>`: the line associates a unique ID to an inode
    - `LOC <UUID> <TIME> <ESCAPED-PATH>`: the line associates a location in time to a file id

The format can be append-only and is easy to parse

Alternative without index
-------------------------

Without an index file synchronized on both parties, a file can be identified by:

- an inode number, but only for the scope of a directory
- a data hash but only at some point in time
- a file name until this is renamed too

It seems impossible to track renames without an index