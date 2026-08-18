import pytest

import main


class TestStripTrailingToolUse:
    def test_removes_trailing_tool_use_block(self):
        messages = [
            {"role": "user", "content": [{"text": "hi"}]},
            {
                "role": "assistant",
                "content": [
                    {"text": "let me check"},
                    {"toolUse": {"name": "search", "input": {}}},
                ],
            },
        ]

        got = main.strip_trailing_tool_use(messages)

        assert got[-1]["content"] == [{"text": "let me check"}]

    def test_pops_message_that_is_entirely_tool_use(self):
        messages = [
            {"role": "user", "content": [{"text": "hi"}]},
            {"role": "assistant", "content": [{"toolUse": {"name": "search", "input": {}}}]},
        ]

        got = main.strip_trailing_tool_use(messages)

        assert len(got) == 1
        assert got[0]["role"] == "user"

    def test_leaves_messages_without_tool_use_untouched(self):
        messages = [{"role": "user", "content": [{"text": "hi"}]}]

        got = main.strip_trailing_tool_use(messages)

        assert got == messages

    def test_rejects_non_list_input(self):
        with pytest.raises(ValueError):
            main.strip_trailing_tool_use("not a list")


class TestExtractPrompt:
    def test_returns_plain_prompt_string(self):
        assert main._extract_prompt({"prompt": "hello"}) == "hello"

    def test_defaults_to_empty_string_when_prompt_missing(self):
        assert main._extract_prompt({}) == ""

    def test_returns_messages_with_trailing_tool_use_stripped(self):
        payload = {
            "messages": [
                {"role": "user", "content": [{"text": "hi"}]},
                {"role": "assistant", "content": [{"toolUse": {"name": "search", "input": {}}}]},
            ]
        }

        got = main._extract_prompt(payload)

        assert len(got) == 1

    def test_converts_tool_results(self):
        payload = {"tool_results": [{"toolUseId": "abc", "content": [{"text": "result"}]}]}

        got = main._extract_prompt(payload)

        assert got == [
            {
                "role": "user",
                "content": [
                    {"toolResult": {"toolUseId": "abc", "status": "success", "content": [{"text": "result"}]}}
                ],
            }
        ]

    def test_rejects_non_dict_payload(self):
        with pytest.raises(ValueError):
            main._extract_prompt("not a dict")

    def test_rejects_non_string_prompt(self):
        with pytest.raises(ValueError):
            main._extract_prompt({"prompt": 123})
