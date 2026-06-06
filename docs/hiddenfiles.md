Mounting hides the original contents of the mount point.

Create:

mkdir hidden-demo
echo "original" > hidden-demo/original.txt

Verify:

ls hidden-demo
original.txt

Create another directory:

mkdir replacement
echo "new" > replacement/new.txt

Now:

sudo mount --bind replacement hidden-demo

Check:

ls hidden-demo

Output:

new.txt

Question:

Where did original.txt go?

Not deleted.

Hidden.

Unmount:

sudo umount hidden-demo

Now:

ls hidden-demo

Output:

original.txt

 the file is not shown because when you do the path resolution it is redirected to different place