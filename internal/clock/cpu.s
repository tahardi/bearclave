#include "textflag.h"

TEXT ·CPUID(SB), NOSPLIT, $0-24
	MOVL	eax+0(FP), AX
	MOVL	ecx+4(FP), CX
	CPUID
	MOVL	AX, reax+8(FP)
	MOVL	BX, rebx+12(FP)
	MOVL	CX, recx+16(FP)
	MOVL	DX, redx+20(FP)
	RET

TEXT ·RDTSC(SB), NOSPLIT, $0-8
	RDTSC
	MOVL	AX, ret+0(FP)
	MOVL	DX, ret+4(FP)
	RET
