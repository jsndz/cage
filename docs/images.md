Docker images are collection of filesystem + metadata

Image has Base image which is read only
and upper layer which is write is done

Container

Writable Layer
--------------
Ubuntu Layer
--------------
Base Layer


Layers are identified by hash

When the container is removed:

docker rm container

the writable layer is deleted.

Writable Layer  ← deleted
Image Layers    ← remain


So,
Image
  ↓
Read-only base filesystem

Container Start
  ↓
Create writable layer

OverlayFS
  ↓
Writable Layer + Image Layers

Container View
  ↓
Single filesystem

