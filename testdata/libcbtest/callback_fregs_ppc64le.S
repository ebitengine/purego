// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

.text
.globl check_callback_fregs
.type check_callback_fregs,@function
check_callback_fregs:
	mflr 0
	std 0,16(1)
	stdu 1,-480(1)
	std 2,24(1)

	// Preserve the caller's registers at 32, sentinels at 176, and callback
	// results at 320.
	.irp r,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31
	stfd \r,32+8*(\r-14)(1)
	.endr
	.irp r,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31
	lis 4,0x1000+\r
	ori 4,4,\r
	std 4,176+8*(\r-14)(1)
	lfd \r,176+8*(\r-14)(1)
	.endr

	mtctr 3
	bctrl
	ld 2,24(1)

	.irp r,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31
	stfd \r,320+8*(\r-14)(1)
	.endr

	li 3,0
	li 4,176
	li 5,320
	li 6,0
.Lcompare_fregs:
	ldx 7,1,4
	ldx 8,1,5
	cmpd 7,8
	beq .Lnext_freg
	li 9,1
	sld 9,9,6
	or 3,3,9
.Lnext_freg:
	addi 4,4,8
	addi 5,5,8
	addi 6,6,1
	cmpdi 6,18
	blt .Lcompare_fregs

	.irp r,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31
	lfd \r,32+8*(\r-14)(1)
	.endr
	addi 1,1,480
	ld 0,16(1)
	mtlr 0
	blr
.size check_callback_fregs,.-check_callback_fregs
.section .note.GNU-stack,"",@progbits
