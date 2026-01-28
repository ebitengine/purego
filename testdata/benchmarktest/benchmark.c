// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

long sum1_c(long a1) { return a1; }

long sum2_c(long a1, long a2) { return a1 + a2; }

long sum3_c(long a1, long a2, long a3) { return a1 + a2 + a3; }

long sum5_c(long a1, long a2, long a3, long a4, long a5) {
  return a1 + a2 + a3 + a4 + a5;
}

long sum10_c(long a1, long a2, long a3, long a4, long a5, long a6, long a7,
             long a8, long a9, long a10) {
  return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10;
}

long sum14_c(long a1, long a2, long a3, long a4, long a5, long a6, long a7, long a8, long a9, long a10, long a11, long a12, long a13, long a14, long a15) {
  return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10 + a11 + a12 + a13 +
         a14 + a15;
}

long sum15_c(long a1, long a2, long a3, long a4, long a5, long a6, long a7, long a8, long a9, long a10, long a11, long a12, long a13, long a14, long a15) {
  return a1 + a2 + a3 + a4 + a5 + a6 + a7 + a8 + a9 + a10 + a11 + a12 + a13 +
         a14 + a15;
}
typedef long (*callback1_t)(long);
long call_callback1(callback1_t cb, long a1) { return cb(a1); }

typedef long (*callback2_t)(long, long);
long call_callback2(callback2_t cb, long a1, long a2) { return cb(a1, a2); }

typedef long (*callback3_t)(long, long, long);
long call_callback3(callback3_t cb, long a1, long a2, long a3) {
  return cb(a1, a2, a3);
}

typedef long (*callback5_t)(long, long, long, long, long);
long call_callback5(callback5_t cb, long a1, long a2, long a3, long a4, long a5) {
  return cb(a1, a2, a3, a4, a5);
}

typedef long (*callback10_t)(long, long, long, long, long, long, long, long, long, long);
long call_callback10(callback10_t cb, long a1, long a2, long a3, long a4, long a5, long a6, long a7, long a8, long a9, long a10) {
  return cb(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10);
}

typedef long (*callback14_t)(long, long, long, long, long, long, long, long, long, long, long, long, long, long);
long call_callback14(callback14_t cb, long a1, long a2, long a3, long a4, long a5, long a6, long a7, long a8, long a9, long a10, long a11, long a12, long a13, long a14) {
  return cb(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14);
}

typedef long (*callback15_t)(long, long, long, long, long, long, long, long, long, long, long, long, long, long, long);
long call_callback15(callback15_t cb, long a1, long a2, long a3, long a4, long a5, long a6, long a7, long a8, long a9, long a10, long a11, long a12, long a13, long a14, long a15) {
  return cb(a1, a2, a3, a4, a5, a6, a7, a8, a9, a10, a11, a12, a13, a14, a15);
}
