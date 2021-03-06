/*
 * Copyright 2017 XLAB d.o.o.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package commitmentzkp

import (
	"math/big"

	"github.com/xlab-si/emmy/crypto/commitments"
	"github.com/xlab-si/emmy/crypto/common"
)

type DFCommitmentOpeningProver struct {
	committer          *commitments.DamgardFujisakiCommitter
	challengeSpaceSize int
	r1                 *big.Int
	r2                 *big.Int
}

func NewDFCommitmentOpeningProver(committer *commitments.DamgardFujisakiCommitter,
	challengeSpaceSize int) *DFCommitmentOpeningProver {
	return &DFCommitmentOpeningProver{
		committer:          committer,
		challengeSpaceSize: challengeSpaceSize,
	}
}

func (prover *DFCommitmentOpeningProver) GetProofRandomData() *big.Int {
	// r1 from [0, T * 2^(NLength + ChallengeSpaceSize))
	nLen := prover.committer.QRSpecialRSA.N.BitLen()
	exp := big.NewInt(int64(nLen + prover.challengeSpaceSize))
	b := new(big.Int).Exp(big.NewInt(2), exp, nil)
	b.Mul(b, prover.committer.T)
	r1 := common.GetRandomInt(b)
	prover.r1 = r1
	// r2 from [0, 2^(B + 2*NLength + ChallengeSpaceSize))
	b = new(big.Int).Exp(big.NewInt(2), big.NewInt(int64(
		prover.committer.B+2*nLen+prover.challengeSpaceSize)), nil)
	r2 := common.GetRandomInt(b)
	prover.r2 = r2
	// G^r1 * H^r2
	proofRandomData := prover.committer.ComputeCommit(r1, r2)
	return proofRandomData
}

func (prover *DFCommitmentOpeningProver) GetProofData(challenge *big.Int) (*big.Int, *big.Int) {
	// s1 = r1 + challenge*a (in Z, not modulo)
	// s2 = r2 + challenge*r (in Z, not modulo)
	a, r := prover.committer.GetDecommitMsg()
	s1 := new(big.Int).Mul(challenge, a)
	s1.Add(s1, prover.r1)
	s2 := new(big.Int).Mul(challenge, r)
	s2.Add(s2, prover.r2)
	return s1, s2
}

type DFCommitmentOpeningVerifier struct {
	receiver           *commitments.DamgardFujisakiReceiver
	challengeSpaceSize int
	challenge          *big.Int
	proofRandomData    *big.Int
}

func NewDFCommitmentOpeningVerifier(receiver *commitments.DamgardFujisakiReceiver,
	challengeSpaceSize int) *DFCommitmentOpeningVerifier {
	return &DFCommitmentOpeningVerifier{
		receiver:           receiver,
		challengeSpaceSize: challengeSpaceSize,
	}
}

func (verifier *DFCommitmentOpeningVerifier) SetProofRandomData(proofRandomData *big.Int) {
	verifier.proofRandomData = proofRandomData
}

func (verifier *DFCommitmentOpeningVerifier) GetChallenge() *big.Int {
	exp := big.NewInt(int64(verifier.challengeSpaceSize))
	b := new(big.Int).Exp(big.NewInt(2), exp, nil)
	challenge := common.GetRandomInt(b)
	verifier.challenge = challenge
	return challenge
}

func (verifier *DFCommitmentOpeningVerifier) Verify(s1, s2 *big.Int) bool {
	// verify proofRandomData * verifier.receiver.Commitment^challenge = G^s1 * H^s2 mod n
	left := verifier.receiver.QRSpecialRSA.Exp(verifier.receiver.Commitment, verifier.challenge)
	left = verifier.receiver.QRSpecialRSA.Mul(verifier.proofRandomData, left)
	right := verifier.receiver.ComputeCommit(s1, s2)
	return left.Cmp(right) == 0
}
