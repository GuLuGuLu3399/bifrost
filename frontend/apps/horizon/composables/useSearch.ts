import { useDebounceFn } from "@vueuse/core";
import { ref, computed, watch } from "vue";
import type { SearchResponse, SuggestResponse } from "@bifrost/shared";
import { useApi } from "./useApi";

/**
 * 搜索 Composable
 * 处理搜索防抖、分页和分面过滤
 */
export const useSearch = () => {
  const query = ref("");
  const pageToken = ref("");
  const pageSize = 10;
  const categoryId = ref<string>();
  const tagId = ref<string>();
  const dateRangeStart = ref<string>();
  const dateRangeEnd = ref<string>();

  // 状态：是否正在搜索
  const isSearching = ref(false);

  // 使用 immediate: false 和 watch: false，由 debouncedSearch 手动触发
  const {
    data: searchResult,
    pending,
    execute: executeSearch,
  } = useApi<SearchResponse>("/v1/search", {
    query: computed(() => ({
      query: query.value,
      "page.page_size": pageSize,
      ...(pageToken.value && { "page.page_token": pageToken.value }),
      include_facets: true,
      ...(categoryId.value && { "filter.category_id": categoryId.value }),
      ...(tagId.value && { "filter.tag_id": tagId.value }),
      ...(dateRangeStart.value && {
        "filter.date_range_start": dateRangeStart.value,
      }),
      ...(dateRangeEnd.value && {
        "filter.date_range_end": dateRangeEnd.value,
      }),
    })),
    immediate: false,
    watch: false,
  });

  // 防抖执行搜索（延迟 300ms）
  const debouncedSearch = useDebounceFn(async () => {
    if (!query.value.trim()) {
      searchResult.value = undefined;
      suggestions.value = undefined;
      isSearching.value = false;
      return;
    }
    isSearching.value = true;
    await Promise.all([executeSearch(), executeSuggest()]);
    isSearching.value = false;
  }, 300);

  // 监听 query 变化，触发防抖搜索
  watch(query, () => {
    debouncedSearch();
  });

  // 搜索建议
  const { data: suggestions, execute: executeSuggest } =
    useApi<SuggestResponse>("/v1/search/suggest", {
      query: computed(() => ({
        prefix: query.value,
        limit: 8,
      })),
      immediate: false,
      watch: false,
    });

  const suggestionList = computed(() => suggestions.value?.suggestions ?? []);

  // 监听搜索建议
  const showSuggestions = computed(() => {
    return query.value.trim() && suggestionList.value.length > 0;
  });

  // 选择建议
  const selectSuggestion = (suggestion: string) => {
    query.value = suggestion;
    void executeSearch();
  };

  return {
    // 输入
    query,
    pageToken,
    categoryId,
    tagId,
    dateRangeStart,
    dateRangeEnd,

    // 输出
    searchResult,
    suggestions: suggestionList,
    showSuggestions,
    pending,
    isSearching,

    // 方法
    selectSuggestion,
  };
};
