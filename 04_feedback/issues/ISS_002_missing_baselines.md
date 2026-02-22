# ISS_002: 비교군(Baseline) 부재

## 메타데이터
- **안건 ID**: ISS_002
- **상태**: open
- **우선순위**: critical
- **출처 피드백**: FB_260209_1500_reviewer
- **대상 섹션**: I (Introduction), IV (Results), V (Discussion)
- **대상 파일**: 초안/260209_1245_실험결과_토론섹션_초안작성_v1.tex
- **생성일**: 2026-02-09
- **해결일**:

## 내용
두 가지 비교군이 모두 누락되어 있다:

### A. 전통 정적 분석 도구 (Lower Bound)
- Introduction에서 "Traditional static analysis tools... are largely blind"라고 주장했으나 실증 데이터 없음
- SonarQube, Semgrep, Go-vet, CodeQL 등을 실제로 돌려서 "0% 탐지" 결과를 Table에 보여줘야 함
- 증거 없는 주장은 학술적으로 부적절

### B. SOTA 클라우드 모델 (Upper Bound)
- GPT-4o, Claude 3.5 Sonnet 등과 비교하여 로컬 sLM의 포지셔닝 근거 필요
- "GPT-4o만큼 잘하는데 비용은 0원" 또는 "프라이버시 보장" 프레이밍이 있어야 가성비 주장이 성립

## 조치 사항
- [ ] SonarQube로 9개 .go 파일 감사 실행 → 결과 기록
- [ ] Semgrep으로 HLF 관련 룰셋 적용 → 결과 기록
- [ ] go vet 실행 → 결과 기록
- [ ] GPT-4o API로 동일 프롬프트 감사 실행 → 결과 기록 (비용 발생)
- [ ] Table에 [Tool/Model × File] 비교 매트릭스 추가
- [ ] 또는 Threats to Validity에서 비교군 미포함 사유를 학술적으로 정당화

## 연결성
- **선행 안건**: ISS_001 (데이터셋 확장 시 비교군도 함께 확장)
- **후행 안건**: 없음
- **관련 마스터가이드**: 논문_마스터가이드/04_방법론/, 논문_마스터가이드/05_실측결과/, 논문_마스터가이드/08_관련연구_참고문헌/
