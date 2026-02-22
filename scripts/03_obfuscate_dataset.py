"""
03_obfuscate_dataset.py
- 02_resources/dataset/ 내 .go 체인코드 파일을 읽어
  사용자 정의 식별자를 난독화한 사본을 02_resources/dataset_obfuscated/에 생성한다.
- 목적: LLM이 변수명/함수명이 아닌 코드 구조와 API 호출 패턴만으로
  취약점을 탐지하는지 검증하기 위한 전처리 단계.
- 난독화 규칙:
    1. 사용자 정의 함수명  -> FuncA, FuncB, ...
    2. 사용자 정의 구조체명 -> StructA, StructB, ...
    3. 사용자 정의 변수명   -> v1, v2, ...  (err, ctx 등 예외)
    4. 구조체 필드명        -> F1, F2, ...  (json 태그도 동기화)
    5. 주석 제거 (// 및 /* */)
    6. HLF API, Go 키워드, 표준 라이브러리 함수, import 경로, 문자열 리터럴 보존
"""

import re
import string
from pathlib import Path
from collections import OrderedDict

# ── 프로젝트 루트 경로 설정 ──────────────────────────────────────────
PROJECT_ROOT = Path(__file__).resolve().parent.parent
DATASET_DIR = PROJECT_ROOT / "02_resources" / "dataset"
OUTPUT_DIR = PROJECT_ROOT / "02_resources" / "dataset_obfuscated"

# ═════════════════════════════════════════════════════════════════════
#  PRESERVE LISTS — 이 이름들은 절대 난독화하지 않는다
# ═════════════════════════════════════════════════════════════════════

# Go 키워드 및 내장 식별자
GO_KEYWORDS = {
    "break", "case", "chan", "const", "continue", "default", "defer",
    "else", "fallthrough", "for", "func", "go", "goto", "if", "import",
    "interface", "map", "package", "range", "return", "select", "struct",
    "switch", "type", "var",
    # 내장 타입
    "bool", "byte", "complex64", "complex128", "error", "float32", "float64",
    "int", "int8", "int16", "int32", "int64", "rune", "string",
    "uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
    # 내장 함수 및 상수
    "append", "cap", "close", "complex", "copy", "delete", "imag", "len",
    "make", "new", "panic", "print", "println", "real", "recover",
    "true", "false", "nil", "iota",
}

# HLF contractapi 관련 보존 이름
HLF_API_NAMES = {
    "contractapi", "Contract", "TransactionContextInterface",
    "NewChaincode", "Start",
    "GetStub", "PutState", "GetState", "DelState",
    "GetStateByRange", "GetStateByPartialCompositeKey",
    "CreateCompositeKey",
    "GetTxTimestamp", "SetEvent",
    "Close", "HasNext", "Next",
}

# Go 표준 라이브러리 패키지 이름 및 함수 (점 호출되는 것들)
STDLIB_NAMES = {
    # 패키지 이름
    "json", "fmt", "time", "math", "rand", "strings", "strconv",
    "sha256", "hex", "sync", "encoding",
    # json
    "Marshal", "Unmarshal",
    # fmt
    "Printf", "Sprintf", "Errorf", "Println",
    # time
    "Now", "Since", "Unix", "Format", "AddDate",
    "RFC3339", "RFC3339Nano",
    # strings
    "Join",
    # strconv
    "Atoi", "Itoa", "ParseFloat", "FormatFloat",
    # crypto/sha256
    "Sum256",
    # encoding/hex
    "EncodeToString",
    # math/rand
    "Intn",
    # math
    "Pow", "Round",
    # sync
    "Mutex", "WaitGroup", "Add", "Done", "Wait", "Lock", "Unlock",
}

# main 함수 및 패키지 선언
MAIN_NAMES = {"main", "package"}

# 작은 루프 / 에러 변수 — 너무 일반적이라 난독화하면 오히려 깨지는 것들
SKIP_VARS = {
    "err", "ctx", "ok", "i", "j", "k", "c", "v", "s", "t", "w",
    "_", "wg", "mu",
}

# 통합 보존 목록
PRESERVE_ALL = GO_KEYWORDS | HLF_API_NAMES | STDLIB_NAMES | MAIN_NAMES | SKIP_VARS

# Protobuf timestamp 필드 접근
PROTOBUF_FIELDS = {"Seconds", "Nanos", "Value", "Key"}
PRESERVE_ALL |= PROTOBUF_FIELDS


# ═════════════════════════════════════════════════════════════════════
#  유틸리티 함수
# ═════════════════════════════════════════════════════════════════════

def _alpha_label(index: int) -> str:
    """0->A, 1->B, ..., 25->Z, 26->AA, 27->AB, ..."""
    result = ""
    while True:
        result = string.ascii_uppercase[index % 26] + result
        index = index // 26 - 1
        if index < 0:
            break
    return result


# ═════════════════════════════════════════════════════════════════════
#  1단계: 문자열 리터럴 및 import 블록 보호
# ═════════════════════════════════════════════════════════════════════

# 플레이스홀더에는 식별자 문자(\w)를 포함하지 않는 특수 구분자를 사용
# 이렇게 하면 \b 경계 매칭에서 절대 식별자로 인식되지 않음
_PH_STR = "\x00\x01#XSTR"     # string placeholder prefix
_PH_IMP = "\x00\x01#XIMP"     # import placeholder prefix
_PH_END = "#\x01\x00"         # placeholder suffix


def _protect_strings_and_imports(code: str):
    """
    문자열 리터럴(쌍따옴표, 백틱)과 import 블록을 플레이스홀더로 치환하여
    후속 정규식 치환에서 건드리지 않도록 보호한다.
    반환: (보호된 코드, 복원 리스트)
    """
    restore_list = []  # [(placeholder, original), ...]
    counter = [0]

    def _make_placeholder(prefix, original):
        idx = counter[0]
        counter[0] += 1
        ph = f"{prefix}{idx}{_PH_END}"
        restore_list.append((ph, original))
        return ph

    def _replace_str(m):
        return _make_placeholder(_PH_STR, m.group(0))

    # 1. 백틱 문자열 (멀티라인 가능) 먼저 보호
    code = re.sub(r'`[^`]*`', _replace_str, code)
    # 2. 쌍따옴표 문자열 (이스케이프된 따옴표 허용) 보호
    code = re.sub(r'"(?:[^"\\]|\\.)*"', _replace_str, code)

    # 3. import 블록 보호
    def _replace_import(m):
        return _make_placeholder(_PH_IMP, m.group(0))

    # import ( ... ) 블록
    code = re.sub(r'import\s*\(.*?\)', _replace_import, code, flags=re.DOTALL)
    # import "..." (단일 임포트)
    code = re.sub(r'import\s+' + re.escape(_PH_STR) + r'\d+' + re.escape(_PH_END),
                  _replace_import, code)

    return code, restore_list


def _restore_protected(code: str, restore_list: list) -> str:
    """보호된 플레이스홀더를 원본으로 복원한다. 역순으로 복원하여 중첩 처리."""
    for ph, original in reversed(restore_list):
        code = code.replace(ph, original)
    return code


# ═════════════════════════════════════════════════════════════════════
#  2단계: 주석 제거
# ═════════════════════════════════════════════════════════════════════

def _remove_comments(code: str) -> str:
    """문자열 리터럴을 보호한 상태에서 주석을 제거한다."""
    # /* ... */ 블록 주석
    code = re.sub(r'/\*.*?\*/', '', code, flags=re.DOTALL)
    # // 라인 주석
    code = re.sub(r'//[^\n]*', '', code)
    # 빈 줄 정리 (3줄 이상 연속 빈 줄 -> 1줄)
    code = re.sub(r'\n{3,}', '\n\n', code)
    return code


# ═════════════════════════════════════════════════════════════════════
#  3단계: 식별자 수집 (구조체, 함수, 필드, 변수)
# ═════════════════════════════════════════════════════════════════════

def _collect_struct_names(code: str) -> list[str]:
    """type StructName struct 패턴에서 구조체 이름 추출."""
    return re.findall(r'\btype\s+([A-Z][A-Za-z0-9_]*)\s+struct\b', code)


def _collect_func_names(code: str) -> list[str]:
    """
    사용자 정의 함수/메서드 이름 수집.
    - func name(...)  형태 (일반 함수)
    - func (r *Type) Name(...)  형태 (메서드)
    main 함수는 제외.
    """
    # 메서드
    methods = re.findall(r'func\s+\(\s*\w+\s+\*?\w+\s*\)\s+(\w+)\s*\(', code)
    # 일반 함수
    plain = re.findall(r'func\s+(\w+)\s*\(', code)

    seen = set()
    result = []
    for name in methods + plain:
        if name in PRESERVE_ALL or name in seen:
            continue
        seen.add(name)
        result.append(name)
    return result


def _collect_struct_fields(code: str) -> list[str]:
    """
    구조체 정의 내부의 필드 이름 수집.
    type ... struct { 블록 내부의 FieldName Type `...` 패턴.
    """
    fields = []
    seen = set()

    # 각 struct 블록을 찾아서 내부 필드 추출
    for m in re.finditer(r'type\s+\w+\s+struct\s*\{([^}]*)\}', code, re.DOTALL):
        body = m.group(1)
        for line in body.strip().split('\n'):
            line = line.strip()
            if not line or line.startswith('//'):
                continue
            # 플레이스홀더로 시작하는 줄은 건너뜀
            if line.startswith('\x00'):
                continue
            # 임베딩 (타입만 있는 줄) 제외: contractapi.Contract 등
            parts = line.split()
            if len(parts) < 2:
                continue
            field_name = parts[0]
            # 점(.)이 있으면 임베딩
            if '.' in field_name:
                continue
            if field_name in PRESERVE_ALL or field_name in seen:
                continue
            # 유효한 Go 식별자
            if re.match(r'^[A-Za-z_]\w*$', field_name):
                seen.add(field_name)
                fields.append(field_name)

    return fields


def _collect_variables(code: str, already_mapped: set[str]) -> list[str]:
    """
    사용자 정의 변수 이름 수집.
    패턴:
      - name := value
      - var name type
      - var name = value
      - for _, name := range
      - name, err := ...
    이미 매핑된 이름과 보존 목록은 제외.
    """
    vars_found = []
    seen = set()

    # := 패턴 (왼쪽의 식별자들)
    for m in re.finditer(r'(\w+(?:\s*,\s*\w+)*)\s*:=', code):
        names_str = m.group(1)
        for name in re.findall(r'\b([a-zA-Z_]\w*)\b', names_str):
            if name not in PRESERVE_ALL and name not in already_mapped and name not in seen:
                seen.add(name)
                vars_found.append(name)

    # var name type / var name = value (함수 내부 -- 들여쓰기 있는 것)
    for m in re.finditer(r'\bvar\s+(\w+)\s+', code):
        name = m.group(1)
        if name not in PRESERVE_ALL and name not in already_mapped and name not in seen:
            seen.add(name)
            vars_found.append(name)

    # 함수 파라미터 이름 (func ... (params) 내부)
    for m in re.finditer(r'func\s+(?:\(\s*\w+\s+\*?\w+\s*\)\s+)?\w+\s*\(([^)]*)\)', code):
        params_str = m.group(1)
        if not params_str.strip():
            continue
        for param in params_str.split(','):
            param = param.strip()
            if not param:
                continue
            parts = param.split()
            if len(parts) >= 2:
                pname = parts[0]
                if pname not in PRESERVE_ALL and pname not in already_mapped and pname not in seen:
                    seen.add(pname)
                    vars_found.append(pname)

    return vars_found


def _collect_const_names(code: str) -> list[str]:
    """const 블록 또는 개별 const 선언에서 이름 수집."""
    consts = []
    seen = set()

    # const ( ... ) 블록 내부
    for m in re.finditer(r'\bconst\s*\(([^)]*)\)', code, re.DOTALL):
        body = m.group(1)
        for line in body.strip().split('\n'):
            line = line.strip()
            if not line or line.startswith('//'):
                continue
            parts = line.split()
            if parts:
                name = parts[0]
                if name not in PRESERVE_ALL and name not in seen and re.match(r'^[a-zA-Z_]\w*$', name):
                    seen.add(name)
                    consts.append(name)

    # 단일 const name = value
    for m in re.finditer(r'\bconst\s+(\w+)\s*=', code):
        name = m.group(1)
        if name not in PRESERVE_ALL and name not in seen:
            seen.add(name)
            consts.append(name)

    return consts


def _collect_global_vars(code: str) -> list[str]:
    """
    함수 외부에 선언된 패키지 레벨 var 수집.
    간단한 휴리스틱: 들여쓰기 없는 var 선언.
    """
    gvars = []
    seen = set()
    for line in code.split('\n'):
        m = re.match(r'^var\s+(\w+)\s+', line)
        if not m:
            m = re.match(r'^var\s+(\w+)\s*=', line)
        if m:
            name = m.group(1)
            if name not in PRESERVE_ALL and name not in seen:
                seen.add(name)
                gvars.append(name)
    return gvars


# ═════════════════════════════════════════════════════════════════════
#  4단계: 난독화 적용
# ═════════════════════════════════════════════════════════════════════

def _build_mapping(names: list[str], prefix: str, label_fn) -> dict[str, str]:
    """이름 -> 난독화 이름 매핑 생성."""
    mapping = OrderedDict()
    for idx, name in enumerate(names):
        mapping[name] = prefix + label_fn(idx)
    return mapping


def _apply_identifier_replacements(code: str, mapping: dict[str, str]) -> str:
    """
    식별자를 매핑에 따라 치환한다.
    단어 경계(\\b)를 사용하여 부분 매칭을 방지한다.
    긴 이름부터 먼저 치환하여 부분 문자열 충돌을 방지한다.
    """
    sorted_names = sorted(mapping.keys(), key=len, reverse=True)
    for name in sorted_names:
        replacement = mapping[name]
        code = re.sub(r'\b' + re.escape(name) + r'\b', replacement, code)
    return code


def _build_field_json_map(code: str, field_names: list[str]) -> dict[str, str]:
    """
    원본 코드에서 각 필드 이름에 대응하는 json 태그 값을 추출한다.
    반환: {원본 필드 이름: json 태그 값}
    """
    field_to_json = {}
    for m in re.finditer(r'type\s+\w+\s+struct\s*\{([^}]*)\}', code, re.DOTALL):
        body = m.group(1)
        for line in body.strip().split('\n'):
            line = line.strip()
            if not line:
                continue
            # FieldName Type `json:"tagValue"`
            tag_match = re.match(r'(\w+)\s+\S+.*?`json:"([^"]+)"`', line)
            if tag_match:
                fname = tag_match.group(1)
                jval = tag_match.group(2)
                if fname in field_names:
                    field_to_json[fname] = jval
    return field_to_json


def _update_json_tags(code: str, field_mapping: dict[str, str],
                      field_json_map: dict[str, str]) -> str:
    """
    구조체 필드의 json 태그를 난독화한다.
    원본 필드의 json 태그 값 -> 난독화된 필드 이름의 소문자 버전으로 교체.

    field_mapping: {원본 필드명 -> 난독화 필드명} (e.g., ID -> F1)
    field_json_map: {원본 필드명 -> 원본 json 태그 값} (e.g., ID -> "id")
    """
    for original_name, obf_name in field_mapping.items():
        json_val = field_json_map.get(original_name)
        if json_val is None:
            continue
        # 난독화된 json 태그: 소문자 변환 (F1 -> f1, F12 -> f12)
        obf_json = obf_name[0].lower() + obf_name[1:]
        # json:"originalValue" -> json:"f1" (보호된 문자열은 아직 플레이스홀더 상태)
        # 하지만 struct 내부의 백틱 태그는 보호 대상(백틱 문자열)에 포함되어 있음
        # 따라서 복원 후에 처리해야 함... 아니면 복원 리스트에서 직접 수정
        pass

    return code


def _update_json_tags_in_restored(code: str, json_replacements: dict[str, str]) -> str:
    """
    복원된 코드에서 json 태그를 직접 치환한다.
    json_replacements: {원본 json 태그 값 -> 난독화 json 태그 값}
    """
    for old_val, new_val in json_replacements.items():
        # `json:"oldVal"` -> `json:"newVal"`
        code = code.replace(f'json:"{old_val}"', f'json:"{new_val}"')
    return code


# ═════════════════════════════════════════════════════════════════════
#  메인 난독화 파이프라인
# ═════════════════════════════════════════════════════════════════════

def obfuscate_go_file(source_code: str) -> tuple[str, dict]:
    """
    Go 소스 코드를 난독화한다.

    반환: (난독화된 코드, 매핑 정보 딕셔너리)
    """
    code = source_code

    # ── (0) 원본 코드에서 필드-json 태그 매핑을 먼저 추출 ─────────────
    #    (주석 제거 및 문자열 보호 전에 추출해야 정확)
    #    주석이 있어도 정규식이 정확한 필드 줄만 매칭하므로 괜찮음

    # ── (1) 문자열 리터럴 및 import 블록 보호 ───────────────────────
    code, restore_list = _protect_strings_and_imports(code)

    # ── (2) 주석 제거 ───────────────────────────────────────────────
    code = _remove_comments(code)

    # ── (3) 식별자 수집 ──────────────────────────────────────────────
    struct_names = _collect_struct_names(code)
    func_names = _collect_func_names(code)
    field_names = _collect_struct_fields(code)
    const_names = _collect_const_names(code)
    global_var_names = _collect_global_vars(code)

    # 원본 코드에서 필드의 json 태그 추출 (보호 전 원본에서)
    field_json_map = _build_field_json_map(source_code, field_names)

    already_mapped = set(struct_names + func_names + field_names + const_names + global_var_names)
    var_names = _collect_variables(code, already_mapped)

    # ── (4) 매핑 생성 ──────────────────────────────────────────────
    struct_map = _build_mapping(struct_names, "Struct", _alpha_label)
    func_map = _build_mapping(func_names, "Func", _alpha_label)
    field_map = _build_mapping(field_names, "F", lambda i: str(i + 1))
    const_map = _build_mapping(const_names, "C", lambda i: str(i + 1))
    gvar_map = _build_mapping(global_var_names, "G", lambda i: str(i + 1))
    var_map = _build_mapping(var_names, "v", lambda i: str(i + 1))

    # ── (5) 치환 적용 ──────────────────────────────────────────────
    combined_map = {}
    combined_map.update(struct_map)
    combined_map.update(func_map)
    combined_map.update(field_map)
    combined_map.update(const_map)
    combined_map.update(gvar_map)
    combined_map.update(var_map)

    code = _apply_identifier_replacements(code, combined_map)

    # ── (6) 문자열 리터럴 및 import 블록 복원 ──────────────────────
    code = _restore_protected(code, restore_list)

    # ── (7) json 태그 난독화 (복원 후 실제 문자열에서 수행) ──────────
    json_replacements = {}
    for original_name, obf_name in field_map.items():
        json_val = field_json_map.get(original_name)
        if json_val is not None:
            obf_json = obf_name[0].lower() + obf_name[1:]
            json_replacements[json_val] = obf_json
    # 긴 태그값부터 교체하여 부분 문자열 충돌 방지
    for old_val in sorted(json_replacements.keys(), key=len, reverse=True):
        new_val = json_replacements[old_val]
        code = code.replace(f'json:"{old_val}"', f'json:"{new_val}"')

    # ── (8) 빈 줄 정리 ────────────────────────────────────────────
    lines = code.split('\n')
    cleaned = []
    prev_empty = False
    for line in lines:
        is_empty = line.strip() == ''
        if is_empty and prev_empty:
            continue
        cleaned.append(line)
        prev_empty = is_empty
    code = '\n'.join(cleaned)

    # 매핑 정보 (디버깅 / 로깅용)
    all_mappings = {
        "structs": struct_map,
        "functions": func_map,
        "fields": field_map,
        "constants": const_map,
        "global_vars": gvar_map,
        "variables": var_map,
    }

    return code, all_mappings


# ═════════════════════════════════════════════════════════════════════
#  메인 실행
# ═════════════════════════════════════════════════════════════════════

def main():
    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)

    go_files = sorted(DATASET_DIR.glob("*.go"))
    if not go_files:
        print(f"[Error] No .go files found in {DATASET_DIR}")
        return

    print(f"{'=' * 65}")
    print(f"  Go Chaincode Obfuscator for LLM Vulnerability Detection Study")
    print(f"{'=' * 65}")
    print(f"  Source : {DATASET_DIR}")
    print(f"  Output : {OUTPUT_DIR}")
    print(f"  Files  : {len(go_files)}")
    print(f"{'=' * 65}\n")

    total_structs = 0
    total_funcs = 0
    total_fields = 0
    total_vars = 0
    total_consts = 0
    total_gvars = 0

    for go_file in go_files:
        source = go_file.read_text(encoding="utf-8")
        obfuscated, mappings = obfuscate_go_file(source)

        out_path = OUTPUT_DIR / go_file.name
        out_path.write_text(obfuscated, encoding="utf-8")

        n_structs = len(mappings["structs"])
        n_funcs = len(mappings["functions"])
        n_fields = len(mappings["fields"])
        n_consts = len(mappings["constants"])
        n_gvars = len(mappings["global_vars"])
        n_vars = len(mappings["variables"])

        total_structs += n_structs
        total_funcs += n_funcs
        total_fields += n_fields
        total_consts += n_consts
        total_gvars += n_gvars
        total_vars += n_vars

        print(f"  [OK] {go_file.name}")
        print(f"       structs={n_structs}  funcs={n_funcs}  fields={n_fields}  "
              f"consts={n_consts}  globals={n_gvars}  vars={n_vars}")

        # 매핑 상세 출력 (디버깅용)
        for category, cat_map in mappings.items():
            if cat_map:
                items = ", ".join(f"{k}->{v}" for k, v in cat_map.items())
                print(f"         {category}: {items}")
        print()

    print(f"{'=' * 65}")
    print(f"  SUMMARY")
    print(f"{'=' * 65}")
    print(f"  Files processed : {len(go_files)}")
    print(f"  Structs renamed : {total_structs}")
    print(f"  Functions renamed: {total_funcs}")
    print(f"  Fields renamed  : {total_fields}")
    print(f"  Constants renamed: {total_consts}")
    print(f"  Globals renamed : {total_gvars}")
    print(f"  Variables renamed: {total_vars}")
    print(f"  Total identifiers: {total_structs + total_funcs + total_fields + total_consts + total_gvars + total_vars}")
    print(f"{'=' * 65}")
    print(f"  Output directory : {OUTPUT_DIR}")
    print(f"{'=' * 65}")


if __name__ == "__main__":
    main()
