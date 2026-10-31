"""
Fail-Soft Tool Chains for Spec-Kit Skills
Each external call has an ordered fallback list. On failure, auto-swap to next tool.
Agent usage remains unchanged.
"""

from typing import List, Callable, Any, Optional, Dict
from dataclasses import dataclass
from abc import ABC, abstractmethod
import time
import logging

logger = logging.getLogger(__name__)


@dataclass
class ToolResult:
    success: bool
    data: Any = None
    error: str = ""
    tool_name: str = ""
    latency_ms: float = 0


class Tool(ABC):
    """Base class for a tool in the fallback chain."""
    
    @property
    @abstractmethod
    def name(self) -> str:
        pass
    
    @abstractmethod
    def execute(self, **kwargs) -> ToolResult:
        pass
    
    def is_available(self) -> bool:
        """Check if tool is available (API key, network, etc.)"""
        return True


class FallbackChain:
    """
    Ordered list of tools for a single action.
    On failure, automatically tries next tool.
    Agent usage unchanged - just call chain.execute().
    """
    
    def __init__(self, action_name: str, tools: List[Tool]):
        self.action_name = action_name
        self.tools = tools
        self.fallback_history: List[Dict] = []
    
    def execute(self, **kwargs) -> ToolResult:
        """Execute with automatic fallback."""
        last_error = ""
        
        for i, tool in enumerate(self.tools):
            if not tool.is_available():
                logger.debug(f"{self.action_name}: {tool.name} not available, skipping")
                continue
            
            start = time.time()
            try:
                result = tool.execute(**kwargs)
                result.latency_ms = (time.time() - start) * 1000
                
                if result.success:
                    if i > 0:
                        logger.info(f"{self.action_name}: Fallback to {tool.name} succeeded (attempt {i+1})")
                        self.fallback_history.append({
                            "tool": tool.name,
                            "attempt": i + 1,
                            "status": "success",
                            "latency_ms": result.latency_ms
                        })
                    return result
                else:
                    last_error = result.error
                    logger.warning(f"{self.action_name}: {tool.name} failed: {result.error}")
                    
            except Exception as e:
                last_error = str(e)
                logger.error(f"{self.action_name}: {tool.name} exception: {e}")
            
            self.fallback_history.append({
                "tool": tool.name,
                "attempt": i + 1,
                "status": "failed",
                "error": last_error,
                "latency_ms": (time.time() - start) * 1000
            })
        
        # All tools failed
        return ToolResult(
            success=False,
            error=f"All {len(self.tools)} tools failed for {self.action_name}. Last error: {last_error}",
            tool_name="fallback_chain_exhausted"
        )


# ============================================================
# PRE-BUILT FALLBACK CHAINS FOR SPEC-KIT SKILLS
# ============================================================

class WebSearchTool(Tool):
    """Web search with multiple backends."""
    
    @property
    def name(self) -> str:
        return self._name
    
    def __init__(self, name: str, search_fn: Callable):
        self._name = name
        self._search_fn = search_fn
    
    def execute(self, query: str, limit: int = 10) -> ToolResult:
        try:
            results = self._search_fn(query, limit)
            return ToolResult(success=True, data=results, tool_name=self._name)
        except Exception as e:
            return ToolResult(success=False, error=str(e), tool_name=self._name)


class WebFetchTool(Tool):
    """Web page fetch with multiple backends."""
    
    @property
    def name(self) -> str:
        return self._name
    
    def __init__(self, name: str, fetch_fn: Callable):
        self._name = name
        self._fetch_fn = fetch_fn
    
    def execute(self, url: str) -> ToolResult:
        try:
            content = self._fetch_fn(url)
            return ToolResult(success=True, data=content, tool_name=self._name)
        except Exception as e:
            return ToolResult(success=False, error=str(e), tool_name=self._name)


# Pre-configured chains for Spec-Kit skills

def create_web_search_chain(primary_search, backup_search, keyless_search) -> FallbackChain:
    """Web search: Primary API -> Backup API -> Keyless fallback."""
    return FallbackChain("web_search", [
        WebSearchTool("primary", primary_search),
        WebSearchTool("backup", backup_search),
        WebSearchTool("keyless", keyless_search),
    ])


def create_web_fetch_chain(jina_reader, exa_reader, browser_fallback) -> FallbackChain:
    """Web fetch: Jina Reader -> Exa -> Browser."""
    return FallbackChain("web_fetch", [
        WebFetchTool("jina_reader", jina_reader),
        WebFetchTool("exa_reader", exa_reader),
        WebFetchTool("browser", browser_fallback),
    ])


def create_llm_chain(primary_llm, backup_llm, keyless_llm) -> FallbackChain:
    """LLM completion: Primary -> Backup -> Keyless (Pollinations)."""
    class LLMTool(Tool):
        def __init__(self, name, fn):
            self._name = name
            self._fn = fn
        
        @property
        def name(self):
            return self._name
        
        def execute(self, prompt: str, **kwargs) -> ToolResult:
            try:
                result = self._fn(prompt, **kwargs)
                return ToolResult(success=True, data=result, tool_name=self._name)
            except Exception as e:
                return ToolResult(success=False, error=str(e), tool_name=self._name)
    
    return FallbackChain("llm_completion", [
        LLMTool("primary", primary_llm),
        LLMTool("backup", backup_llm),
        LLMTool("keyless", keyless_llm),
    ])


def create_youtube_visuals_chain(pexels, pixabay, nasa, openverse, ai_gen) -> FallbackChain:
    """Visuals: Pexels -> Pixabay -> NASA -> Openverse -> AI-generated."""
    class VisualTool(Tool):
        def __init__(self, name, fn):
            self._name = name
            self._fn = fn
        
        @property
        def name(self):
            return self._name
        
        def execute(self, query: str, count: int = 5) -> ToolResult:
            try:
                results = self._fn(query, count)
                # Ensure each result has credit
                for r in results:
                    if "credit" not in r:
                        r["credit"] = f"Source: {self._name}"
                return ToolResult(success=True, data=results, tool_name=self._name)
            except Exception as e:
                return ToolResult(success=False, error=str(e), tool_name=self._name)
    
    return FallbackChain("visuals", [
        VisualTool("pexels", pexels),
        VisualTool("pixabay", pixabay),
        VisualTool("nasa", nasa),
        VisualTool("openverse", openverse),
        VisualTool("ai_generated", ai_gen),
    ])


def create_transcription_chain(groq_whisper, openai_whisper, local_whisper) -> FallbackChain:
    """Transcription: Groq Whisper -> OpenAI Whisper -> Local Whisper."""
    class TranscribeTool(Tool):
        def __init__(self, name, fn):
            self._name = name
            self._fn = fn
        
        @property
        def name(self):
            return self._name
        
        def execute(self, audio_path: str) -> ToolResult:
            try:
                result = self._fn(audio_path)
                return ToolResult(success=True, data=result, tool_name=self._name)
            except Exception as e:
                return ToolResult(success=False, error=str(e), tool_name=self._name)
    
    return FallbackChain("transcription", [
        TranscribeTool("groq_whisper", groq_whisper),
        TranscribeTool("openai_whisper", openai_whisper),
        TranscribeTool("local_whisper", local_whisper),
    ])


def create_thumbnail_chain(primary_style, fallback_styles) -> FallbackChain:
    """Thumbnail generation with 4 styles as fallbacks."""
    class ThumbnailTool(Tool):
        def __init__(self, name, style_fn):
            self._name = name
            self._style_fn = style_fn
        
        @property
        def name(self):
            return self._name
        
        def execute(self, title: str, format: str = "long") -> ToolResult:
            try:
                result = self._style_fn(title, format)
                if result.get("size_mb", 0) > 2:
                    return ToolResult(success=False, error="Thumbnail > 2MB", tool_name=self._name)
                return ToolResult(success=True, data=result, tool_name=self._name)
            except Exception as e:
                return ToolResult(success=False, error=str(e), tool_name=self._name)
    
    tools = [ThumbnailTool("primary", primary_style)]
    for style in fallback_styles:
        tools.append(ThumbnailTool(style, style))
    
    return FallbackChain("thumbnail", tools)