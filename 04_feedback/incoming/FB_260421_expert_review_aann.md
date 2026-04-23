# FB_260421 — AANN 2026 Expert Review (원문)

- **수신일**: 2026-04-21
- **출처**: AANN 2026 Expert Reviewer
- **대상 논문**: v51 (AANN_2026_SungmoonPark.pdf)
- **Paper No.**: M7VNBFDSWP
- **Order No.**: 226041801092265613

---

## 원문 (English)

### 5. Conclusion

This paper addresses the critical problem of detecting silent nondeterminism in Hyperledger Fabric chaincode which leads to endorsement failures without explicit runtime errors. This paper finds that the Qwen2.5-Coder-7B model achieves high classification accuracy and matches prompt engineered cloud models while operating offline on consumer hardware. Although the results are promising the small scale of the ground truth dataset and the potential for naming cue bias suggest a need for further revision to ensure the findings are generalizable.

### 6. Comment

1. This paper relies on existing pre-trained models and standard prompting techniques without proposing any architectural innovations or specialized fine-tuning for blockchain security tasks.
2. This paper assumes that file-level classification is sufficient for auditing when industrial applications require precise line-level localization to be practically useful for developers.
3. This paper bases its primary conclusions on a very small dataset of fifteen files which lacks the statistical power to represent diverse production environments.
4. This paper should standardize the naming of the Gemini models to match official versioning and ensure consistency across all tables and textual descriptions.
5. This paper should expand the references to include recent studies on large language models applied specifically to Go language security and static analysis.

---

## 번역문 (Korean)

### 5. 결론 (Reviewer Summary)

본 논문은 명시적인 런타임 오류 없이 승인 실패를 초래하는 Hyperledger Fabric 체인코드의 숨겨진 비결정성을 탐지하는 중요한 문제를 다룹니다. 본 연구 결과, Qwen2.5-Coder-7B 모델은 높은 분류 정확도를 달성하며, 일반 소비자용 하드웨어에서 오프라인으로 작동하면서도 클라우드 기반 모델과 유사한 성능을 보이는 것으로 나타났습니다. 결과는 고무적이지만, 정답 데이터셋의 규모가 작고 명명 단서 편향 가능성이 존재하므로, 연구 결과의 일반화 가능성을 확보하기 위해 추가적인 수정이 필요합니다.

### 6. 댓글 (Comments)

1. 본 논문은 블록체인 보안 작업을 위한 아키텍처 혁신이나 특수 미세 조정 기법을 제안하지 않고 기존의 사전 학습된 모델과 표준 프롬프트 기법에 의존합니다.
2. 본 논문은 산업 응용 프로그램에서 개발자에게 실질적인 유용성을 제공하기 위해 정확한 줄 단위 현지화가 요구되는 경우 파일 수준 분류만으로도 감사가 가능하다고 가정합니다.
3. 본 논문의 주요 결론은 다양한 운영 환경을 대표할 통계적 검정력이 부족한 매우 작은 15개 파일 데이터셋에 기반합니다.
4. 본 논문은 Gemini 모델의 명명법을 공식 버전 관리 체계에 맞춰 표준화하고 모든 표와 텍스트 설명에서 일관성을 유지해야 합니다.
5. 본 논문은 Go 언어 보안 및 정적 분석에 특화된 대규모 언어 모델에 대한 최근 연구를 참고 문헌에 포함해야 합니다.
