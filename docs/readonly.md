
  ## How Readonly Rootfs Works

  The overlay filesystem you already have is the key. You have two modes:

   Mode                                         | Overlay Setup                                            | What the container sees
  ----------------------------------------------|----------------------------------------------------------|------------------------------------------------
   Default (writable)                           |  lowerdir=rootfs, upperdir=upper, workdir=work  → merged | Writable filesystem (writes go to upper layer)
   Readonly                                     | Mount  lowerdir  directly, no upper/work                 | Truly immutable rootfs

  ### The Mechanism

  With  --readonly , instead of mounting a full overlay (lower + upper + work), you just bind-mount the lower dir as the root and remount it read-only. No
  upper layer means no writes are possible.

  In your filesystem.go:

    // MountReadOnly bind-mounts the rootfs as read-only (no overlay upper layer).
    func MountReadOnly(lower, merged string) error {
        if err := os.MkdirAll(merged, 0755); err != nil {
            return err
        }
        // Bind mount the lower dir
        if err := unix.Mount(lower, merged, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
            return err
        }
        // Remount as read-only
        return unix.Mount("", merged, "", unix.MS_REMOUNT|unix.MS_BIND|unix.MS_RDONLY, "")
    }
    
  ### How It Flows Through Your Code
  1. Flag in main.go — add  --readonly :

    readOnly := flag.Bool("readonly", false, "mount rootfs as read-only")
    
  2. Thread it through — the  readOnly  bool needs to reach  InitContainer . You already have an  initPayload  pipe pattern in runtime.go, so add it
  there:

    type initPayload struct {
        ContainerIP     string                   `json:"container_ip"`
        SecurityConfig  *security.SecurityConfig `json:"security_config"`
        ApparmorProfile string                   `json:"apparmor_profile"`
        ReadOnly        bool                     `json:"read_only"`
    }
    
  3. Branch in runtime.go:

    // After receiving payload from pipe:
    if payload.ReadOnly {
        // No overlay — bind-mount lower as read-only
        if err := filesystem.MountReadOnly(lowerlayer, merged); err != nil {
            panic(err)
        }
    } else {
        // Default: full overlay with writable upper layer
        if err := filesystem.MountOverlay(lowerlayer, upperlayer, workdir, merged); err != nil {
            panic(err)
        }
    }

  4. Cleanup also needs to branch — readonly mode has no overlay to clean, just an unmount:

    // In StartContainer, after cmd.Wait():
    if readOnly {
        unix.Unmount("/tmp/overlay/merged", 0)
    } else {
        filesystem.CleanOverlay("/tmp/overlay/merged", "/tmp/overlay")
    }

  ### Why Lower Dir as Root by Default Makes Sense

  Your default keeps the overlay (writable). The  --readonly  flag skips the overlay entirely and uses  lowerdir  directly. This is exactly how Docker does
  it —  docker run --read-only  makes the rootfs immutable while still allowing writes to explicitly mounted tmpfs/volumes.

  ### One thing to consider

  With a truly readonly rootfs, things like writing  /etc/resolv.conf  (which you do runtime.go) will fail. You'll need to either:

  • Write  resolv.conf  before the readonly remount (won't work since it's the base image)
  • Mount a small  tmpfs  on  /etc  or  /tmp  even in readonly mode so DNS config and temp files work

  Docker solves this by always mounting tmpfs on  /tmp  and bind-mounting a generated  resolv.conf .