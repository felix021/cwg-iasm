package x86_64

import (
    `fmt`
    `os`
    `os/exec`
    `testing`

    `github.com/cloudwego/iasm/asm`
    `github.com/cloudwego/iasm/obj`
    `github.com/davecgh/go-spew/spew`
    `github.com/stretchr/testify/require`
)

func TestAssembler_Assemble(t *testing.T) {
    p := asm.GetArch("x86_64").CreateAssembler()
    e := p.Assemble(`
.org 0x08000000
.entry start

start:
    movl    $1, %edi
    leaq    msg(%rip), %rsi
    movl    $13, %edx
    movl    $0x02000004, %eax
    syscall
    xorl    %edi, %edi
    movl    $0x02000001, %eax
    syscall

msg:
    .ascii "hello, world\n"
`)
    require.NoError(t, e)
    code := p.Code()
    base := p.Base()
    entry := p.Entry()
    spew.Dump(code)
    fmt.Printf("Image Base  : %#x\n", base)
    fmt.Printf("Entry Point : %#x\n", entry)
    err := obj.CurrentOS.CompileAndLink("/tmp/iasm-out", obj.GetArch("x86_64"), code, base, entry)
    require.NoError(t, err)
    println("Saved to /tmp/iasm-out")
    out, err := exec.Command("/tmp/iasm-out").Output()
    require.NoError(t, err)
    spew.Dump(out)
    require.Equal(t, []byte("hello, world\n"), out)
    err = os.Remove("/tmp/iasm-out")
    require.NoError(t, err)
}

func TestAssembler_RIPRelative(t *testing.T) {
    p := asm.GetArch("x86_64").CreateAssembler()
    e := p.Assemble(`leaq 0x1b(%rip), %rsi`)
    require.NoError(t, e)
    spew.Dump(p.Code())
    require.Equal(t, []byte { 0x48, 0x8d, 0x35, 0x1b, 0x00, 0x00, 0x00 }, p.Code())
}

func TestAssembler_AbsoluteAddressing(t *testing.T) {
    p := asm.GetArch("x86_64").CreateAssembler()
    e := p.Assemble(`
movq 0x1234, %rbx
movq %rcx, 0x1234
`)
    require.NoError(t, e)
    spew.Dump(p.Code())
    require.Equal(t, []byte {
        0x48, 0x8b, 0x1c, 0x25, 0x34, 0x12, 0x00, 0x00,
        0x48, 0x89, 0x0c, 0x25, 0x34, 0x12, 0x00, 0x00,
    }, p.Code())
}

func TestAssembler_LockPrefix(t *testing.T) {
    p := asm.GetArch("x86_64").CreateAssembler()
    e := p.Assemble(`lock cmpxchgq %r9, (%rbx)`)
    require.NoError(t, e)
    spew.Dump(p.Code())
    require.Equal(t, []byte { 0xf0, 0x4c, 0x0f, 0xb1, 0x0b }, p.Code())
}

func TestAssembler_SegmentOverride(t *testing.T) {
    p := asm.GetArch("x86_64").CreateAssembler()
    e := p.Assemble(`
movq gs:0x30, %rcx
movq gs:10(%rax), %rcx
movq fs:(%r9), %rcx
movq %rcx, gs:0x30
movq %rcx, gs:10(%rax)
movq %rcx, fs:(%r9)
`)
    require.NoError(t, e)
    spew.Dump(p.Code())
    require.Equal(t, []byte {
        0x65, 0x48, 0x8b, 0x0c, 0x25, 0x30, 0x00, 0x00, 0x00,
        0x65, 0x48, 0x8b, 0x48, 0x0a,
        0x64, 0x49, 0x8b, 0x09,
        0x65, 0x48, 0x89, 0x0c, 0x25, 0x30, 0x00, 0x00, 0x00,
        0x65, 0x48, 0x89, 0x48, 0x0a,
        0x64, 0x49, 0x89, 0x09,
    }, p.Code())
}
