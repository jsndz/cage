mount event is something that changes the mount table

lets say in a mount namespace you add something to the mount table

the change is not only visible to that namespace but also to the host namespace

this is because mount is marked as share

Shared Mounts

Think:

Host
   ↔
Container

Mount changes propagate both ways.

Host mounts something
      ↓
Container sees it

Container mounts something
      ↓
Host sees it


Now we need to change them to private mount

there are also slave mount which propgate one way