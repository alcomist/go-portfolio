# Real GO Lang Project Skeleton Structure and Codes

## Overview

그동안 작업해본 GO 프로젝트에서 하나의 공통 기반 라이브러리 코드를 바탕으로 여러 개의 cli 실행 파일 및 web api 기반 코드 모음입니다.

* [Github](http://github.com/alcomist)

## Go Directories

### `/build`

실제 cli 폴더에 있는 GO 파일로부터 컴파일된 실행파일이 저장되는 곳입니다.

### `/cli`

tasker에 command arguments를 적용할 수 있도록 wrapping한 struct들이 있는 곳이며 실제 컴파일되어 build 폴더에 들어가게 되는

main 패키지들이 모여 있는 곳입니다.

### `/config`

각종 설정 파일들의 예시 파일을 두는 곳입니다.

### `/internal`

프로젝트 전체에서 쓰이는 각종 편의 함수 들을 모아둔 곳입니다. 

상수 모음, 로그 시스템, 해쉬 생성코드, 슬랙 연동 등 다양한 코드를 이곳에 모아서 task 패키지에 있는 tasker들에서 참조하여

해당 코드를 사용합니다.

### `/task`

실제 문제 해결을 위한 tasker 모음입니다. 각 폴더는 해당 tasker의 성격으로 기초해서 분류하였고 성격에 따라 하위 tasker가 해당 내용에 대한 처리를 담당합니다.

### `/test`

internal에 있는 library의 테스트 코드 모음입니다. 여기에서 모든 기초 library의 정합성을 담당합니다.