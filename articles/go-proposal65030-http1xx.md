---
title: "[Go言語Proposalを読む] net/http: Clientが1xxレスポンスの受信回数を制限しないようにする"
emoji: "😽"
type: "tech" # tech: 技術記事 / idea: アイデア
topics: [Go]
published: false
---

# [Go言語Proposalを読む] シリーズについて

- このシリーズでは[Go言語](https://github.com/golang/go)に提案されたProposalのうち、Acceptedとなったものについて1つ取り上げて解説します。
  - 「Acceptedになったもの」とは、将来のGoで実装され、リリースされる予定であることを意味します。
  - AcceptされたProposalの結論はissueのタイトルと異なっている場合があるので、記事のタイトルは「最終的な結論を筆者が日本語で書いたもの」になっています。
- Go言語については一定のハンズオン経験がある読者を仮定しています。
- 次のようなことがわかるように書くことを目指しています:
  - Proposalの前提となる知識
    - Proposalの対象(また、対象となっているGoパッケージ・型・関数の役割)
    - Proposalの背景となっているソフトウェアエンジニアリングやコンピュータサイエンスのトピック
  - Proposalが解決しようとする問題
  - Proposalが問題を解決する手段
  - 検討された他の解決手段
- 記事執筆のモチベーションは、特に「Proposalの背景となっているソフトウェアエンジニアリングやコンピュータサイエンスのトピック」について学習することにあります。つまり、前提知識部分がメインディッシュです。

# 基本資料

- https://github.com/golang/go/issues/65035 Proposal
  - https://github.com/golang/go/issues/65035#issuecomment-2433286714 結論部分
- https://datatracker.ietf.org/doc/draft-ietf-httpbis-resumable-upload/

# Proposalの対象

この記事で扱うProposalは https://github.com/golang/go/issues/65035 です。タイトルは

> net/http: customize limit on number of 1xx responses

で、

> net/http: 1xxレスポンスの受信回数制限をカスタムする

というような意味ですが、最終的な結論はこれとは異なります。

`net/http`パッケージの[`*http.Client`](https://pkg.go.dev/net/http#Client)型が対象です。

# Proposalが解決しようとする問題

HTTPのレスポンスにはレスポンスコードが含まれています。

そのうち1xxレスポンスつまり100番台のレスポンスは[情報レスポンス](https://developer.mozilla.org/ja/docs/Web/HTTP/Status#%E6%83%85%E5%A0%B1%E3%83%AC%E3%82%B9%E3%83%9D%E3%83%B3%E3%82%B9)として使われます。

この1xxレスポンスは、最終的なレスポンス(200など)の前に任意回数返されることがあります。1回も返されないかもしれないし、複数回返されることもあるということです。ところが現在のhttp.Clientは1回のリクエストにつき5回までしか1xxレスポンスを受け取らず、6回目の1xxレスポンスを受け取った時には`*Client.Do`メソッドがエラーを返します。

これはHTTPの仕様に従っていないのでHTTPクライアントとしての[バグである](https://github.com/golang/go/issues/65035#issuecomment-1894104055)ようです。

:::message

根拠となる HTTP仕様はおそらく https://httpwg.org/specs/rfc9110.html#overview.of.status.codes にある次の部分と思われます。

> A client MUST be able to parse one or more 1xx responses received prior to a final response, even if the client does not expect one. 

:::

# Proposalが問題を解決する手段

そこで、

- 5回という恣意的な制限をなくす
- その代わり、デフォルトでは、受信したすべての1xxレスポンスヘッダーの合計サイズが、所定のサイズを超えたらエラーにする
- この挙動はClient設定でカスタムできる(httptrace.ClientTrace型のGot1xxResponseを使う)

実装もすでに始まっているようです(HTTP/2):
https://github.com/golang/net/commit/4783315416d92ff3d4664762748bd21776b42b98



## 背景: Resumable Uploads for HTTP 

Proposalで挙げられている他の事情として、IETFに提案されている https://datatracker.ietf.org/doc/draft-ietf-httpbis-resumable-upload/ の存在があります。

:::message

こちらについて、日本語記事の https://asnokaze.hatenablog.com/entry/2022/02/28/010418 が参考になりました。

:::

この提案の中で、サーバー側からクライアントへアップロードの進捗を繰り返し通知する機能を検討しているので、5回という制限があるとそれがうまくいかない、という事情もあるようです。

:::message
https://datatracker.ietf.org/doc/draft-ietf-httpbis-resumable-upload/ の中に該当する記述があるか探したのですが見つけられませんでした。検討中ということでまだdraftにも反映されていない内容なのかもしれません。わかる方がいたら教えてください。
:::


# 検討された他の解決手段

元々のproposalは、5回という恣意的な回数制限をhttp.Clientのフィールドで上書きできるようにすることを提案していました。

# そのほかProposalから得られる知識

## `net/http/httptrace`パッケージ

https://pkg.go.dev/net/http/httptrace

パッケージドキュメントにあるように、HTTP Clientからのリクエストによって起こるイベントをトレースするための機能を提供します。

https://pkg.go.dev/net/http/httptrace#ClientTrace 型が中心となる型で、さまざまなイベントに対するイベントハンドラー的なコールバック関数をフィールドとして設定しておき、これを`context.Context`に詰めて`http.Request`型に設定するという使い方のようです。

:::message
パッケージの名称からするとトレースとかデバッグ・ロギングに使うものなのかなという印象を持ちますが、この記事で読んだProposalによって、`Got1xxResponse`を設定すると1xxレスポンスに対するデフォルト挙動が変わるので、必ずしもtraceという目的に使うパッケージではなくなるような気がしますね。
:::

# この記事へのフィードバックについて

- この記事についてフィードバックやご意見がある場合、[GitHubリポジトリ](https://github.com/nobishino/zenn)にissueかPRを立てていただけると助かります。
  - ZennのコメントよりもGitHub上でのやり取りが好ましいです。
  - GitHub上ではissueを立てずにいきなりPRを立てても大丈夫です。
- Go言語のProposalはソフトウェア開発やコンピュータサイエンスの多くの分野に関わるため、筆者の関連知識が十分ではない場合が多いです。自信のない内容はそれとわかるように書くつもりですが、思い込みなどで誤った内容を断言してしまうこともあるかと思います。その際はぜひフィードバックをお願いいたします。

