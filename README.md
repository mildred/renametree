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

