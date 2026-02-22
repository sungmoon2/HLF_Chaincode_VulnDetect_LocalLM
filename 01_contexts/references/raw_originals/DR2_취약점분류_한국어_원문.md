# [원문 전문] Deep Research 2: 취약점 분류 체계 (한국어)
> 원본 제목: "하이퍼레저 패브릭 체인코드 보안성 검증을 위한 LLM 평가 데이터셋 구축 및 취약점 심층 분석 보고서"
> 생성일: 2026-02-09
> 도구: Gemini Deep Research
> 분석 요약: 03_vulnerability_taxonomy_kr.md

(이 파일의 전문 내용은 분석 요약본 03_vulnerability_taxonomy_kr.md에 코드 블록을 포함하여 원문에 가깝게 구조화되어 있습니다. 03 파일이 사실상 원문 전문 역할을 수행합니다.)

## 원문 구조
- 서론 (Introduction)
- 1. 하이퍼레저 패브릭 아키텍처와 결정론적 실행의 중요성
  - 1.1 실행-순서화-검증 (Execute-Order-Validate) 모델의 이해
  - 1.2 결정론적 실행(Deterministic Execution)의 필수성
  - 1.3 체인코드 컨테이너의 생명주기와 전역 변수
- 2. LLM 평가를 위한 취약점 데이터셋 상세 분석 (Dataset Generation)
  - Case 1: 비결정론적 실행 (Non-deterministic Execution) — time.Now()
  - Case 2: 전역 변수 오남용 (Global Variable Misuse) — var globalCounter
  - Case 3: 고루틴 동시성 문제 (Goroutine Concurrency) — go func()
  - Case 4: 맵 순회 비결정론 (Map Iteration Non-determinism) — range map
  - Case 5: 쿼리 반복자 미종료 (Unclosed Query Iterator) — GetStateByRange
- 3. 요약 및 결론 (Conclusion)
- 표 1: 하이퍼레저 패브릭 체인코드 취약점 요약 (5행)

각 Case에 shim 기반 Go 코드 예제 + VULNERABILITY 주석 + 분석 및 LLM 평가 포인트 포함.
