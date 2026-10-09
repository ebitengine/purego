// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

.text
.globl check_callback_vregs
.type check_callback_vregs,@function
check_callback_vregs:
	mflr 0
	std 0,16(1)
	stdu 1,-608(1)
	std 2,24(1)

	li 4,32
	stvx 20,1,4
	li 4,48
	stvx 21,1,4
	li 4,64
	stvx 22,1,4
	li 4,80
	stvx 23,1,4
	li 4,96
	stvx 24,1,4
	li 4,112
	stvx 25,1,4
	li 4,128
	stvx 26,1,4
	li 4,144
	stvx 27,1,4
	li 4,160
	stvx 28,1,4
	li 4,176
	stvx 29,1,4
	li 4,192
	stvx 30,1,4
	li 4,208
	stvx 31,1,4

	vspltisw 20,1
	vspltisw 21,2
	vspltisw 22,3
	vspltisw 23,4
	vspltisw 24,5
	vspltisw 25,6
	vspltisw 26,7
	vspltisw 27,8
	vspltisw 28,9
	vspltisw 29,10
	vspltisw 30,11
	vspltisw 31,12

	li 4,224
	stvx 20,1,4
	li 4,240
	stvx 21,1,4
	li 4,256
	stvx 22,1,4
	li 4,272
	stvx 23,1,4
	li 4,288
	stvx 24,1,4
	li 4,304
	stvx 25,1,4
	li 4,320
	stvx 26,1,4
	li 4,336
	stvx 27,1,4
	li 4,352
	stvx 28,1,4
	li 4,368
	stvx 29,1,4
	li 4,384
	stvx 30,1,4
	li 4,400
	stvx 31,1,4

	mr 12,3
	mtctr 3
	bctrl
	ld 2,24(1)

	li 4,416
	stvx 20,1,4
	li 4,432
	stvx 21,1,4
	li 4,448
	stvx 22,1,4
	li 4,464
	stvx 23,1,4
	li 4,480
	stvx 24,1,4
	li 4,496
	stvx 25,1,4
	li 4,512
	stvx 26,1,4
	li 4,528
	stvx 27,1,4
	li 4,544
	stvx 28,1,4
	li 4,560
	stvx 29,1,4
	li 4,576
	stvx 30,1,4
	li 4,592
	stvx 31,1,4

	li 3,0
	li 4,224
	li 5,416
	li 6,0
.Lcompare:
	ldx 7,1,4
	ldx 8,1,5
	cmpd 7,8
	beq .Lcompare_second_word
	addi 4,4,16
	addi 5,5,16
	b .Lclobbered
.Lcompare_second_word:
	addi 4,4,8
	addi 5,5,8
	ldx 7,1,4
	ldx 8,1,5
	addi 4,4,8
	addi 5,5,8
	cmpd 7,8
	beq .Lnext
.Lclobbered:
	li 9,1
	sld 9,9,6
	or 3,3,9
.Lnext:
	addi 6,6,1
	cmpdi 6,12
	blt .Lcompare
	li 4,32
	lvx 20,1,4
	li 4,48
	lvx 21,1,4
	li 4,64
	lvx 22,1,4
	li 4,80
	lvx 23,1,4
	li 4,96
	lvx 24,1,4
	li 4,112
	lvx 25,1,4
	li 4,128
	lvx 26,1,4
	li 4,144
	lvx 27,1,4
	li 4,160
	lvx 28,1,4
	li 4,176
	lvx 29,1,4
	li 4,192
	lvx 30,1,4
	li 4,208
	lvx 31,1,4

	addi 1,1,608
	ld 0,16(1)
	mtlr 0
	blr
.size check_callback_vregs,.-check_callback_vregs
